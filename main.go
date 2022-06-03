package main

import (
	"flag"
	"fmt"

	"github.com/davidalvarez305/chico/actions"
)

var (
	purchase bool
	launch   bool
	crawl    bool
	deploy   bool
	domain   string
	db       string
	siteName string
)

func init() {
	flag.BoolVar(&purchase, "purchase", false, "Command for purchasing a domain")
	flag.BoolVar(&launch, "launch", false, "Launch a new server instance.")
	flag.BoolVar(&crawl, "crawl", false, "Crawl products for a specific website.")
	flag.BoolVar(&deploy, "deploy", false, "Deploy changes to a specific project.")
	flag.StringVar(&domain, "d", "chico.com", "Define the domain to be purchased.")
	flag.StringVar(&db, "db", "", "Define this project's database.")
	flag.StringVar(&siteName, "s", "", "Define this project's database.")
}

func main() {
	flag.Parse()

	if purchase {
		actions.PurchaseDomain(domain)
		fmt.Printf("Domain purchased successfully.")
	}

	if launch {
		actions.LaunchServer(domain, db, siteName)
		fmt.Printf("Server launched successfully.")
	}
}
