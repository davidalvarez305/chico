package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/davidalvarez305/chico/actions"
	"github.com/joho/godotenv"
)

var (
	purchase  bool
	launch    bool
	crawl     bool
	deploy    bool
	domain    string
	db        string
	siteName  string
	project   string
	syncFiles bool
	all       bool
)

func init() {
	flag.BoolVar(&purchase, "purchase", false, "Command for purchasing a domain")
	flag.BoolVar(&launch, "launch", false, "Launch a new server instance.")
	flag.BoolVar(&crawl, "crawl", false, "Crawl products for a specific website.")
	flag.BoolVar(&deploy, "deploy", false, "Deploy changes to a specific project.")
	flag.StringVar(&domain, "d", "chico.com", "Define the domain to be purchased.")
	flag.StringVar(&db, "db", "", "Define this project's database.")
	flag.StringVar(&siteName, "s", "", "Define this project's database.")
	flag.StringVar(&project, "p", "", "Name of the project to be deployed.")
	flag.BoolVar(&syncFiles, "sync", false, "Sync all projects.")
	flag.BoolVar(&all, "all", false, "Deploy all projects.")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	flag.Parse()

	if purchase {
		actions.PurchaseDomain(domain)
		fmt.Printf("Domain purchased successfully.\n")
	}

	if launch {
		actions.LaunchServer(domain, db, siteName)
		fmt.Printf("Server launched successfully.\n")
	}

	if deploy {
		username := os.Getenv("GITHUB_USER")
		actions.Deploy(all, username, project)
		fmt.Printf("Deployed successfully.\n")
	}

	if syncFiles {
		actions.SyncFiles()
	}
}
