package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/davidalvarez305/chico/utils"
)

func main() {
	utils.ParseArguments(os.Args[1:])

	var domainFlag string
	flag.StringVar(&domainFlag, "d", "chico.com", "Define the domain to be purchased.")

	flag.Parse()
	fmt.Println(domainFlag)
}
