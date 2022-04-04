package main

import (
	"fmt"
	"os"

	"github.com/bartek/gotunnel/tunnel"
	"gopkg.in/yaml.v3"
)

func main() {
	f, err := os.OpenFile("config.yml", os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var config tunnel.Config

	if err := yaml.NewDecoder(f).Decode(&config); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	c := make(chan bool)

	for _, dt := range config.Tunnels {
		// Get the identity
		// FIXME: Sanity check the path somewhere, maybe on yaml decode. Since
		// we panic otherwise
		var ident string
		for _, i := range config.Identity {
			if i.Name == dt.Identity {
				ident = i.Path
				break
			}
		}

		fmt.Println(dt.Target, dt.Local, dt.Remote, ident)
		t := tunnel.New(
			dt.Target,
			tunnel.PEMFile(ident),
			dt.Local,
			dt.Remote,
		)
		go t.Start() // err = ..
		//if err != nil {
		//	fmt.Println(err)
		//}
	}

	<-c

}
