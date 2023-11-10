package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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

		// pull
		pull = flag.Bool("pull", false, "git pull all internal folders")
	)

	flag.Parse()

	current, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if *pull {
		err = filepath.Walk(".",
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					return nil
				}
				fld := filepath.Join(path, ".git")
				if _, err := os.Stat(fld); err != nil {
					return nil
				}
				cmd := exec.Command("git", "pull")
				cmd.Dir = filepath.Join(current, path)
				fmt.Println("Dir:", cmd.Dir)
				out, err := cmd.Output()
				if err != nil {
					fmt.Println("Error:", path, err)
					err = nil
				} else {
					fmt.Println(string(out))
				}
				return filepath.SkipDir
			})
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println("Done.")
		return
	}

	rs, err := gitup(*userName)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(-1)
	}
	for _, repos := range rs {
		fmt.Fprintf(os.Stdout, "%s\n", repos)
		if *clone {
			cmd := exec.Command("git", "clone", repos)
			cmd.Dir = current
			_, err := cmd.Output()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			} else {
				fmt.Fprintf(os.Stdout, "%s ... Done\n", repos)
			}
		}
	}
}

func gitup(userName string) (rs []string, err error) {
	if userName == "" {
		err = fmt.Errorf("user name is empty")
		return
	}

	for page := 1; page < 100; page++ {
		url := fmt.Sprintf(
			"https://api.github.com/users/%s/repos?page=%d&per_page=100",
			userName, page)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			err = fmt.Errorf("cannot get request: %v", err)
			return nil, err
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			err = fmt.Errorf("cannot get responce: %v", err)
			return nil, err
		}

		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			err = fmt.Errorf("cannot read body: %v", err)
			return nil, err
		}

		lines := strings.Split(string(body), ",")
		found := false
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
				return nil, err
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
				return nil, err
			}
			rs = append(rs, repos)
			found = true
		}
		if !found {
			break
		}
	}
	sort.Strings(rs)
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
