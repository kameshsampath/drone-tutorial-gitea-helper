package commands

import (
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"path"
	"strings"

	"code.gitea.io/sdk/gitea"
)

//newGiteaClient creates new Gitea Client
func (opts *WorkshopOptions) newGiteaClient() (*gitea.Client, error) {
	c, err := gitea.NewClient(opts.GiteaURL)
	c.SetBasicAuth(opts.GiteaAdminUser, opts.GiteaAdminPassword)
	if err != nil {
		return nil, err
	}
	return c, nil
}

//randomHex generates and returns a random 16 digit Hex value
func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

//repoNameFromURL extracts the lastpath segment from the Git Repo url
func repoNameFromURL(repoUrl string) (string, error) {
	myUrl, err := url.Parse(repoUrl)
	if err != nil {
		return "", err
	}

	repoName := path.Base(myUrl.Path)

	if strings.HasSuffix(repoName, ".git") {
		repoName = strings.TrimSuffix(repoName, ".git")
	}

	return repoName, nil
}
