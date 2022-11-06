package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func main() {
	var (
		// get json from github. Example:
		// https://api.github.com/users/Konstantin8105/repos?page=$PAGE&per_page=1000
		//                              --------------
		//                                 User name
		userName = flag.String("u", "", "set the user name of the github repository")

		// clone
		clone = flag.Bool("clone", false, "clone all repositories")
	)

	flag.Parse()

	rs, err := gitup(*userName)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(-1)
	}
	for _, repos := range rs {
		fmt.Fprintf(os.Stdout, "%s\n", repos)
		if *clone {
			_, err := exec.Command("git", "clone", repos).Output()
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func gitup(userName string) (rs []string, err error) {
	if userName == "" {
		err = fmt.Errorf("user name is empty")
		return
	}

	url := fmt.Sprintf(
		"https://api.github.com/users/%s/repos?page=$PAGE&per_page=1000",
		userName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		err = fmt.Errorf("cannot get request: %v", err)
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("cannot get responce: %v", err)
		return
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		err = fmt.Errorf("cannot read body: %v", err)
		return
	}

	lines := strings.Split(string(body), ",")
	for i := range lines {
		// find lines like
		//"clone_url": "https://github.com/Konstantin8105/c4go.git",
		if !strings.Contains(lines[i], "\"clone_url\"") {
			continue
		}
		// parse
		index := strings.Index(lines[i], ":")
		if index < 0 {
			err = fmt.Errorf("cannot found letter `:` in: %v", lines[i])
			return
		}
		repos := lines[i][index+1:]
		repos = strings.Replace(repos, "\"", "", -1)
		repos = strings.Replace(repos, ",", "", -1)
		repos = strings.TrimSpace(repos)
		err = isOk(repos)
		if err != nil {
			err = fmt.Errorf("cannot clone repository `%v`: %v",
				repos,
				err)
			return
		}
		rs = append(rs, repos)
	}

	return
}

func isOk(repos string) (err error) {
	// pattern of repository name is
	// https://github.com/Konstantin8105/c4go.git
	if !strings.HasPrefix(repos, "https://github.com/") ||
		!strings.HasSuffix(repos, ".git") {
		return fmt.Errorf("repository name if not in pattern")
	}
	return nil
}
