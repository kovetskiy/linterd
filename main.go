package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docopt/docopt-go"
)

var (
	version = "1.0"
)

const (
	usage = `

Usage:
	linterd [options]

Options:
	-c <config>       Use specified configuration file [default: /etc/linterd]
`
)

type Server struct {
	api      *StashAPI
	lintArgs []string
}

func main() {
	args, err := docopt.Parse(usage, nil, true, version, true, true)
	if err != nil {
		panic(err)
	}

	config, err := getConfig(args["-c"].(string))
	if err != nil {
		log.Fatalf("can't load configuration file: %s", err)
	}

	stashAPI, err := getStashAPI(
		config.StashHost, config.StashUser, config.StashPassword,
	)
	if err != nil {
		log.Fatal("can't create stash api resource: %s", err)
	}

	server := &Server{
		api:      stashAPI,
		lintArgs: config.LintArgs,
	}

	log.Printf("Listening on %s", config.ListenAddress)
	err = http.ListenAndServe(config.ListenAddress, server)
	if err != nil {
		log.Fatal(err)
	}
}

func (server *Server) ServeHTTP(
	response http.ResponseWriter, request *http.Request,
) {
	requestURL := strings.TrimPrefix(request.URL.Path, "/")

	cloneURL, branch, err := server.getCloneURLAndBranch(requestURL)
	if err != nil {
		log.Println(err)
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("cloning repository '%s'", cloneURL)

	gopathDirectory, repositoryDirectory, err := cloneRepository(cloneURL)
	if err != nil {
		log.Printf("can't clone repository '%s': %s", cloneURL, err)
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("checkout repository '%s' to '%s'", repositoryDirectory, branch)

	err = checkoutRepository(repositoryDirectory, branch)
	if err != nil {
		log.Printf(
			"can't checkout %s (%s) to '%s': %s",
			repositoryDirectory, cloneURL, branch, err,
		)
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("running go get for %s", repositoryDirectory)

	err = goget(gopathDirectory, repositoryDirectory)
	if err != nil {
		log.Printf(
			"go get for %s with gopath=%s (%s) for branch %s failed: %s ",
			gopathDirectory, repositoryDirectory, cloneURL, branch, err,
		)
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf(
		"linting %s with args %s",
		repositoryDirectory, server.lintArgs,
	)

	output := lint(repositoryDirectory, server.lintArgs)

	response.Write([]byte(output))
}

func (server *Server) getCloneURLAndBranch(
	url string,
) (cloneURL string, branch string, err error) {
	matches := reStashURL.FindStringSubmatch(url)
	if len(matches) > 0 {
		var (
			project     = matches[4]
			repository  = matches[5]
			pullRequest = matches[6]
		)

		return server.api.GetPullRequestInfo(
			project, repository, pullRequest,
		)
	}

	return url, "refs/heads/master", nil
}

func cloneRepository(url string) (string, string, error) {
	gopathDirectory, err := ioutil.TempDir(os.TempDir(), "linterd_")
	if err != nil {
		return "", "", fmt.Errorf("can't create temp directory: %s", err)
	}

	repositoryDirectory := filepath.Join(
		gopathDirectory, "src", "linterd-target",
	)

	cmd := exec.Command(
		"git", "clone", url, repositoryDirectory,
	)
	_, err = execute(cmd)
	if err != nil {
		return "", "", err
	}

	return gopathDirectory, repositoryDirectory, nil
}

func checkoutRepository(repositoryDirectory string, branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = repositoryDirectory
	_, err := execute(cmd)
	return err
}

func goget(gopathDirectory, repositoryDirectory string) error {
	cmd := exec.Command("go", "get", "-v")
	cmd.Dir = repositoryDirectory

	environ := os.Environ()
	for index, env := range environ {
		if strings.HasPrefix(env, "GOPATH=") {
			environ[index] = ""
		}
	}

	cmd.Env = append(
		environ, "GOPATH="+gopathDirectory, "GO15VENDOREXPERIMENT=1",
	)

	fmt.Printf("XXXXXX main.go:165: cmd.Env: %#v\n", cmd.Env)
	_, err := execute(cmd)
	return err
}

func lint(repositoryDirectory string, args []string) string {
	cmd := exec.Command("gometalinter", args...)
	cmd.Dir = repositoryDirectory
	output, _ := execute(cmd)
	return output
}
