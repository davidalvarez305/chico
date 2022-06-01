package controller

import (
	"github.com/davidalvarez305/chico/actions"
	"github.com/davidalvarez305/chico/types"
)

func Controller(opts types.Command) {
	if opts.Command == "purchase" {
		actions.PurchaseDomain(opts.Value)
	}
}
