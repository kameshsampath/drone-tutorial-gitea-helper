package commands

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"

	"code.gitea.io/sdk/gitea"
	yamlv2 "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	cwd string
	err error
)

func init() {
	cwd, err = os.Getwd()
	if err != nil {
		panic(err)
	}
}

func TestSetupWorkshop(t *testing.T) {
	workshopConfigFile := path.Join(cwd, "testdata", "workshop.yaml")
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}
	if kubeconfig == "" {
		t.Fatal("Unable to get and set kubeconfig")
	}
	rootCmd := NewRootCommand()
	rootCmd.SetArgs([]string{"setup-workshop", "-f", workshopConfigFile, "-k", kubeconfig, "-v", "debug"})
	rootCmd.Execute()

	//Query and check if users are created
	var workshopOpts WorkshopOptions

	b, err := ioutil.ReadFile(workshopConfigFile)
	if err != nil {
		t.Fatalf("%v", v)
	}

	err = yamlv2.Unmarshal(b, &workshopOpts)
	if err != nil {
		t.Fatalf("%v", v)
	}

	c, err := workshopOpts.newGiteaClient()
	if err != nil {
		t.Fatalf("%v", v)
	}

	users, _, err := c.AdminListUsers(gitea.AdminListUsersOptions{})
	if err != nil {
		t.Fatalf("%v", v)
	}

	//factor admin user
	if len(users) != workshopOpts.GiteaUsers.To+1 {
		t.Fatalf("Expecting %d users but got %d", workshopOpts.GiteaUsers.To, len(users))
	}

	expectedUsers := []string{"user-01", "user-02"}
	actualUsers := make([]string, len(users)-1)
	for _, u := range users {
		if u.IsAdmin || u.UserName == "demo" {
			continue
		}
		actualUsers = append(actualUsers, u.UserName)
	}

	if reflect.DeepEqual(expectedUsers, actualUsers) {
		t.Errorf("Expecting users %v but got %v", expectedUsers, actualUsers)
	}

	expectedOAuthApps := []string{"demo-oauth-user-01", "demo-oauth-user-02"}
	oAuthApps, _, err := c.ListOauth2(gitea.ListOauth2Option{})

	if err != nil {
		t.Fatalf("%v", v)
	}

	actualOAuthApps := make([]string, len(oAuthApps))

	for _, o := range oAuthApps {
		actualOAuthApps = append(actualOAuthApps, o.Name)
	}
	if reflect.DeepEqual(expectedUsers, actualUsers) {
		t.Errorf("Expecting oAuthAps %v but got %v", expectedOAuthApps, actualOAuthApps)
	}

	//Testbed teardown

	//Delete the Secrets
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Logf("Error getting kubernetes client %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Logf("Error getting building clientset %v", err)
	}

	ns := workshopOpts.GiteaUsers.Namespace
	if ns == "" {
		ns = "default"
	}

	for _, a := range expectedOAuthApps {
		clientset.CoreV1().Secrets(ns).Delete(context.TODO(), fmt.Sprintf("%s-secret", a), metav1.DeleteOptions{})
	}

	for _, u := range users {
		t.Logf("\nDeleting user and their repos %v", u)

		if !u.IsAdmin || u.UserName != "demo" {
			for _, repoURL := range workshopOpts.GiteaUsers.Repos {
				repoName, err := repoNameFromURL(repoURL)
				if err != nil {
					t.Logf("Error finding repo name %s", err)
					continue
				}
				if _, err := c.DeleteRepo(u.UserName, repoName); err != nil {
					t.Logf("Error deleting repo name %s for user %s", err, u.UserName)
					continue
				}
			}
			if _, err := c.AdminDeleteUser(u.UserName); err != nil {
				t.Logf("Error deleting user %v, %v", u, err)
			}
		}
	}

}
