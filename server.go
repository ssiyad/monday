package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"

	sigar "github.com/cloudfoundry/gosigar"
)

type procStat struct {
	PID       int    `json:"pid"`
	PPID      int    `json:"ppid"`
	Mem       uint64 `json:"mem"`
	TimeStart uint64 `json:"timeStart"`
	TimeTotal uint64 `json:"timeTotal"`
	Command   string `json:"command"`
}

type memStat struct {
	Total      uint64 `json:"total"`
	Used       uint64 `json:"used"`
	Free       uint64 `json:"free"`
	ActualUsed uint64 `json:"actualUsed"`
	ActualFree uint64 `json:"actualFree"`
}

type swapStat struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
	Free  uint64 `json:"free"`
}

type fsStat struct {
	Dev   string `json:"dev"`
	Dir   string `json:"dir"`
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
	Avail uint64 `json:"avail"`
}

type stat struct {
	Uptime         string     `json:"uptime"`
	LoadAvgOne     float64    `json:"loadAvgOne"`
	LoadAvgFive    float64    `json:"loadAvgFive"`
	LoadAvgFifteen float64    `json:"loadAvgFifteen"`
	ProcsTopFive   []procStat `json:"procsTopFive"`
	Mem            memStat    `json:"mem"`
	Swap           swapStat   `json:"swap"`
	FsList         []fsStat   `json:"fsList"`
}

type httpHandler struct{}

func (h httpHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	s := &stat{}

	concreteSigar := sigar.ConcreteSigar{}

	uptime := sigar.Uptime{}
	uptime.Get()
	s.Uptime = uptime.Format()

	avg, err := concreteSigar.GetLoadAverage()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}
	s.LoadAvgOne = avg.One
	s.LoadAvgFive = avg.Five
	s.LoadAvgFifteen = avg.Fifteen

	pids := sigar.ProcList{}
	pids.Get()

	for _, pid := range pids.List {
		state := sigar.ProcState{}
		mem := sigar.ProcMem{}
		time := sigar.ProcTime{}

		if err := state.Get(pid); err != nil {
			continue
		}
		if err := mem.Get(pid); err != nil {
			continue
		}
		if err := time.Get(pid); err != nil {
			continue
		}

		s.ProcsTopFive = append(s.ProcsTopFive, procStat{
			pid,
			state.Ppid,
			mem.Resident,
			time.StartTime,
			time.Total,
			state.Name,
		})
	}

	sort.Slice(s.ProcsTopFive, func(i, j int) bool {
		return s.ProcsTopFive[i].Mem > s.ProcsTopFive[j].Mem
	})
	s.ProcsTopFive = s.ProcsTopFive[:5]

	mem := sigar.Mem{}
	mem.Get()
	swap := sigar.Swap{}
	swap.Get()

	s.Mem = memStat{
		mem.Total,
		mem.Used,
		mem.Free,
		mem.ActualUsed,
		mem.ActualFree,
	}
	s.Swap = swapStat{
		swap.Total,
		swap.Used,
		swap.Free,
	}

	fsList := sigar.FileSystemList{}
	fsList.Get()

	for _, fs := range fsList.List {
		dir_name := fs.DirName

		usage := sigar.FileSystemUsage{}

		usage.Get(dir_name)

		s.FsList = append(s.FsList, fsStat{
			fs.DevName,
			fs.DirName,
			usage.Total,
			usage.Used,
			usage.Avail,
		})
	}

	j, err := json.Marshal(s)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}

	res.Header().Add("Content-Type", "application/json")
	fmt.Fprint(res, string(j))
}

// Server is the main function for server
func Server() {
	fmt.Println("starting server on port 4278")
	handler := httpHandler{}
	fmt.Fprint(os.Stderr, http.ListenAndServe(":4278", handler))
}
