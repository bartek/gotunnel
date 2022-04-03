package main

import (
	"fmt"

	"github.com/bartek/gotunnel/tunnel"
)

func main() {
	t := tunnel.New(
		"ec2-54-71-219-77.us-west-2.compute.amazonaws.com:22",
		tunnel.PEMFile("/Users/bartek/.ssh/lumen-dogfood.pem"),
		"127.0.0.1:3001",
		"grafana.graph:3000",
	)
	err := t.Start()
	if err != nil {
		fmt.Println(err)
	}
}
