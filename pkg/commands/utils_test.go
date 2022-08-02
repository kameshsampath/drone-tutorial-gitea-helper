package commands

import "testing"

func TestRepoNameFromURL(t *testing.T) {
	repoURL := "https://github.com/kameshsampath/jar-stack"
	repoName, err := repoNameFromURL(repoURL)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if repoName != "jar-stack" {
		t.Errorf("Expecting 'jar-stack' but got %s", repoName)
	}
}

func TestRepoNameFromURLWithExt(t *testing.T) {
	repoURL := "https://github.com/kameshsampath/jar-stack.git"
	repoName, err := repoNameFromURL(repoURL)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if repoName != "jar-stack" {
		t.Errorf("Expecting 'jar-stack' but got %s", repoName)
	}
}
