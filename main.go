package main

import (
	"flag"
	"fmt"
	"log"
	"os/user"

	"github.com/davidalvarez305/chico/actions"
	"github.com/joho/godotenv"
)

var (
	purchase  bool
	launch    bool
	crawl     bool
	deploy    bool
	replicate bool
	domain    string
	db        string
	siteName  string
	keyword   string
	project   string
	syncFiles bool
)

func init() {
	flag.BoolVar(&purchase, "purchase", false, "Command for purchasing a domain")
	flag.BoolVar(&launch, "launch", false, "Launch a new server instance.")
	flag.BoolVar(&crawl, "crawl", false, "Crawl products for a specific website.")
	flag.BoolVar(&deploy, "deploy", false, "Deploy changes to a specific project.")
	flag.BoolVar(&syncFiles, "sync", false, "Sync all projects.")
	flag.BoolVar(&replicate, "replicate", false, "Upload DB SQL files.")
	flag.StringVar(&domain, "d", "chico.com", "Define the domain to be purchased.")
	flag.StringVar(&db, "db", "", "Define this project's database.")
	flag.StringVar(&siteName, "s", "", "Define this project's database.")
	flag.StringVar(&project, "p", "", "Name of the project to be deployed.")
	flag.StringVar(&keyword, "kw", "", "Seed keyword for crawling.")
}

func main() {
	dir, err := user.Current()

	if err != nil {
		log.Fatal("Error getting current home directory.")
	}

	env := dir.HomeDir + "/chico/.env"

	fmt.Println(env)

	err = godotenv.Load(env)
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	flag.Parse()

	if purchase {
		if domain == "" {
			log.Fatal("Missing d flag.")
		}
		actions.PurchaseDomain(domain)
		fmt.Printf("Domain purchased successfully.\n")
	}

	if launch {
		if domain == "" {
			log.Fatal("Missing d flag.")
		}
		if db == "" {
			log.Fatal("Missing db flag.")
		}
		if siteName == "" {
			log.Fatal("Missing s flag.")
		}
		actions.LaunchServer(domain, db, siteName)
		fmt.Printf("Server launched successfully.\n")
	}

	if deploy {
		if project == "" {
			log.Fatal("Missing p flag.")
		}
		actions.Deploy(project)
		fmt.Printf("Deployed successfully.\n")
	}

	if replicate {
		if project == "" {
			log.Fatal("Missing p flag.")
		}
		actions.Replicate(project)
		fmt.Printf("DB SQl files uploaded & deployed successfully.\n")
	}

	if syncFiles {
		actions.SyncFiles()
	}

	if crawl {
		if keyword == "" {
			log.Fatal("Missing kw flag.")
		}
		actions.Crawl(keyword)
	}
}
