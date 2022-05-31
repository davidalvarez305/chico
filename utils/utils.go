package utils

import (
	"fmt"

	"github.com/davidalvarez305/chico/types"
)

func ParseArguments(args []string) {
	var options types.Options
	for i := 0; i < len(args); i++ {
		options.Command = args[i]
	}
	fmt.Printf("%+v", options)
}
