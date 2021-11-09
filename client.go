package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type server struct {
	Name string `yaml:"name"`
	IP   string `yaml:"ip"`
	Port int32  `yaml:"port"`
}

type config struct {
	Servers []server `yaml:"servers"`
}

func printServer(s server) {
	ip, port := s.IP, s.Port
	if port == 0 {
		port = 4278
	}

	url := fmt.Sprintf("http://%s:%d", ip, port)

	r, err := http.Get(url)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}

	j := stat{}
	err = json.Unmarshal(b, &j)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}

	fmt.Fprintf(os.Stdout, "%s\t%s\n", s.Name, s.IP)

	fmt.Fprintf(os.Stdout, "Uptime\t%s\n\n", j.Uptime)

	fmt.Fprintf(os.Stdout, "%8s %8s %8s %8s\n", "", "1 Min", "5 Min", "15 Min")
	fmt.Fprintf(os.Stdout, "%-8s %8.2f %8.2f %8.2f\n\n", "Load", j.LoadAvgOne, j.LoadAvgFive, j.LoadAvgFifteen)

	fmt.Fprintf(os.Stdout, "%8s %8s %8s %8s\n", "", "Total", "Used", "Free")
	fmt.Fprintf(os.Stdout, "%-8s %8.2f %8.2f %8.2f\n", "Memory", bytesToGiB(j.Mem.Total), bytesToGiB(j.Mem.Used), bytesToGiB(j.Mem.Free))
	fmt.Fprintf(os.Stdout, "%-8s %8.2f %8.2f %8.2f\n\n", "Swap", bytesToGiB(j.Swap.Total), bytesToGiB(j.Swap.Used), bytesToGiB(j.Swap.Free))

	for _, v := range j.FsList {
		fmt.Fprintf(os.Stdout, "%s\n", v.Dir)
		fmt.Fprintf(os.Stdout, "%-8s %-8s %-8s\n", "Total", "Used", "Avail")
		fmt.Fprintf(os.Stdout, "%-8.2f %-8.2f %-8.2f\n\n", bytesToGiB(v.Total), bytesToGiB(v.Used), bytesToGiB(v.Avail))
	}

	for _, v := range j.ProcsTopFive {
		fmt.Fprintf(os.Stdout, "%s: %d\n", "pid", v.PID)
		fmt.Fprintf(os.Stdout, "%s: %.2f\n", "mem", bytesToGiB(v.Mem))
		fmt.Fprintf(os.Stdout, "%s: %s\n\n", "cmd", v.Command)
	}
}

func Client() {
	confHome := os.Getenv("XDG_CONFIG_HOME")
	if confHome == "" {
		dir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
		}
		confHome = filepath.Join(dir, ".config")
	}
	confHome = filepath.Join(confHome, "monday")
	confPath := filepath.Join(confHome, "config.yml")

	_, err := os.Stat(confHome)
	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(confHome, os.ModePerm)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
		}
	}

	s, err := os.Stat(confPath)
	if errors.Is(err, os.ErrNotExist) {
		_, err = os.OpenFile(confPath, os.O_RDONLY|os.O_CREATE, os.FileMode(0666))
		if err != nil {
			fmt.Fprint(os.Stderr, err)
		}
	}

	f, err := os.Open(confPath)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}

	y := make([]byte, s.Size())
	_, err = f.Read(y)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}

	c := config{}

	err = yaml.Unmarshal(y, &c)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}

	for _, v := range c.Servers {
		printServer(v)
	}
}

func bytesToGiB(b uint64) float64 {
	return float64(b) / 1.074e+9
}
