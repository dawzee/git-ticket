package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/thought-machine/gonduit"
	"github.com/thought-machine/gonduit/core"
)

// arcConfig reflects the JSON arcanist configuration file
type arcConfig struct {
	PhabUrl string `json:"phabricator.uri"`
}

// getApiToken returns the Phabricator API token from the repository config
func getApiToken() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to get the current working directory: %q", err)
	}

	repo, err := NewGitRepoNoInit(cwd)
	if err == ErrNotARepo {
		return "", fmt.Errorf("must be run from within a git repo")
	}

	var apiToken string
	if apiToken, err = repo.LocalConfig().ReadString("daedalean.taskmgr-api-token"); err != nil {
		if apiToken, err = repo.GlobalConfig().ReadString("daedalean.taskmgr-api-token"); err != nil {
			phabUrl, err := getPhabUrl()
			if err != nil {
				return "", err
			}
			msg := `No Phabricator API token set. Please go to
	%s/settings/user/<YOUR_USERNAME_HERE>/page/apitokens/
click on <Generate API Token>, and then paste the token into this command
	git config --global --replace-all daedalean.taskmgr-api-token <PASTE_TOKEN_HERE>`
			return "", fmt.Errorf(msg, phabUrl)
		}
	}

	return apiToken, nil
}

// getPhabUrl returns the Phabricator Url held in the arcconfig file
func getPhabUrl() (string, error) {
	dat, err := ioutil.ReadFile(".arcconfig")
	if err != nil {
		return "", err
	}

	var config arcConfig

	err = json.Unmarshal(dat, &config)
	if err != nil {
		return "", err
	}

	if config.PhabUrl == "" {
		return "", errors.New(".arcconfig missing phabricator.uri")
	}

	return config.PhabUrl, nil
}

// GetPhabClient returns the connection ready to be queried. Must be called
// within a git repo which has a .arconfig file containing the phabricator.uri
// field and the Phabricator conduit API token set in the git config
// daedalean.taskmgr-api-token.
func GetPhabClient() (*gonduit.Conn, error) {
	apiToken, err := getApiToken()
	if err != nil {
		return nil, err
	}

	phabUrl, err := getPhabUrl()
	if err != nil {
		return nil, err
	}

	return gonduit.Dial(phabUrl, &core.ClientOptions{APIToken: apiToken})
}
