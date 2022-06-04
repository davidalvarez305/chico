package actions

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/davidalvarez305/chico/types"
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

func filterReport(projectName string, repos []types.GithubJSONResponse) string {
	var project string
	for i := 0; i < len(repos); i++ {
		if strings.Contains(repos[i].CloneURL, projectName) {
			project = repos[i].CloneURL
		}
	}

	return project
}

func Deploy(userName string, projectName string) {
	r, err := getRepos(userName)

	if err != nil {
		log.Fatal("Failed getting repos %v\n", err)
	}

	deploymentProject := filterReport(projectName, r)
	fmt.Printf("Project: %s\n", deploymentProject)
}
