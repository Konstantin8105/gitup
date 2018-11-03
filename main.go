package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func main() {
	// get json from github. Example:
	// https://api.github.com/users/Konstantin8105/repos?page=$PAGE&per_page=1000
	//                              --------------
	//                                 User name
	var userName = flag.String("u", "", "set the user name of the github repository")

	flag.Parse()

	err := gitup(*userName)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(-1)
	}
}

func gitup(userName string) (err error) {
	if userName == "" {
		return fmt.Errorf("user name is empty")
	}

	url := fmt.Sprintf(
		"https://api.github.com/users/%s/repos?page=$PAGE&per_page=1000",
		userName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("cannot get request: %v", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot get responce: %v", err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("cannot read body: %v", err)
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
			return fmt.Errorf("cannot found letter `:` in: %v", lines[i])
		}
		repos := lines[i][index+1:]
		repos = strings.Replace(repos, "\"", "", -1)
		repos = strings.Replace(repos, ",", "", -1)
		repos = strings.TrimSpace(repos)
		err = clone(repos)
		if err != nil {
			return fmt.Errorf("cannot clone repository `%v`: %v",
				repos,
				err)
		}
	}

	return nil
}

func clone(repos string) (err error) {
	// pattern of repository name is
	// https://github.com/Konstantin8105/c4go.git
	if !strings.HasPrefix(repos, "https://github.com/") ||
		!strings.HasSuffix(repos, ".git") {
		return fmt.Errorf("repository name if not in pattern")
	}

	fmt.Println(repos)
	return nil
}
