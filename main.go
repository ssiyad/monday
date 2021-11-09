package main

import (
	"flag"
)

func main() {
	flag.Parse()
	switch flag.Arg(0) {
	case "server":
		Server()
	default:
		Client()
	}
}
