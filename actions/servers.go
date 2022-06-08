package actions

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/davidalvarez305/chico/types"
	"github.com/davidalvarez305/chico/utils"
)

func getRepos(userName string) ([]types.GithubJSONResponse, error) {
	var repo []types.GithubJSONResponse
	repos_url := fmt.Sprintf("https://api.github.com/users/%s/repos", userName)

	resp, err := http.Get(repos_url)

	if err != nil {
		return repo, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return repo, err
	}

	json.Unmarshal(body, &repo)

	return repo, nil
}

func filterReport(projectName string) types.Project {
	var project types.Project
	var projects []types.Project

	body, err := os.ReadFile("projects.json")

	if err != nil {
		log.Fatal("Failed getting repos %v\n", err)
	}

	json.Unmarshal(body, &projects)
	for i := 0; i < len(projects); i++ {
		if strings.Contains(projects[i].Project, projectName) {
			project = projects[i]
		}
	}

	return project
}

func Deploy(all bool, userName, projectName string) {
	var projects []types.Project

	body, err := os.ReadFile("projects.json")

	if err != nil {
		log.Fatal("Failed getting repos %v\n", err)
	}

	json.Unmarshal(body, &projects)

	if all {
		for i := 0; i < len(projects); i++ {
			utils.DeployProject(projects[i])
		}
	} else {
		deploymentProject := filterReport(projectName)
		utils.DeployProject(deploymentProject)
	}
}

func SyncFiles() {
	var projects []types.Project

	body, err := os.ReadFile("projects.json")

	if err != nil {
		log.Fatal("Failed getting repos %v\n", err)
	}

	json.Unmarshal(body, &projects)

	for i := 0; i < len(projects); i++ {
		utils.SecureCopy(projects[i].Key, projects[i].IP, projects[i].Project)
	}

	fmt.Printf("Finalized syncing folders.")
}
