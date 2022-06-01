package main

import (
	"flag"
	"fmt"

	"github.com/davidalvarez305/chico/controller"
	"github.com/davidalvarez305/chico/types"
)

func main() {

	var command string
	flag.StringVar(&command, "purchase", "purchase", "Command for purchasing a domain")

	var domain string
	flag.StringVar(&domain, "d", "chico.com", "Define the domain to be purchased.")

	options := types.Command{
		Command: command,
		Value:   domain,
	}

	fmt.Printf("%+v", options)
	flag.Parse()
	controller.Controller(options)
}
