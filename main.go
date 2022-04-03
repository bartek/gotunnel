package main

import (
	"fmt"

	"github.com/bartek/gotunnel/tunnel"
)

func main() {
	t := tunnel.New(
		"ubuntu@ec2-54-71-219-77.us-west-2.compute.amazonaws.com",
		tunnel.PEMFile("/Users/bartekc/.ssh/lumen-dogfood.pem"),
		"127.0.0.1:19072",
		"vespa.graph:19071",
	)
	err := t.Start()
	if err != nil {
		fmt.Println(err)
	}
}
