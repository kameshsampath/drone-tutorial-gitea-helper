package commands

import (
	"fmt"
	"path/filepath"
	"testing"

	"k8s.io/client-go/util/homedir"
)

func TestCreateOAuthAppWithDefaults(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.SetArgs([]string{"oauthapp", "-a", "defaults", "-v", "debug"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("%v", err)
	}
}

func TestCreateOAuthAppUpdate(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.SetArgs([]string{"oauthapp", "-a", "defaults", "-v", "debug"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("%v", err)
	}
}

func TestCreateOAuthAppWithK8sSecret(t *testing.T) {
	rootCmd := NewRootCommand()
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}
	if kubeconfig == "" {
		t.Fatal("Unable to get and set kubeconfig")
	}
	rootCmd.SetArgs([]string{"oauthapp", "-a", "k8s-secret", "-s", "-n", "default", "-k", kubeconfig, "-v", "debug"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("%v", err)
	}
}

func TestCreateOAuthAppWithNoName(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.SetArgs([]string{"oauthapp", "-v", "debug"})
	eErr := fmt.Errorf("required flag(s) \"app-name\" not set")
	if err := rootCmd.Execute(); err != nil {
		if err.Error() != eErr.Error() {
			t.Fatalf("Expecting error %v but got %v", eErr, err)
		}
	}
}

func TestCreateOAuthAppWithK8sSecretNoNamespace(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.SetArgs([]string{"oauthapp", "-a", "k8s-secret-no-ns", "-s", "-v", "debug"})
	eErr := fmt.Errorf("require namespace to create the k8s-secret-no-ns secret")
	if err := rootCmd.Execute(); err != nil {
		if err.Error() != eErr.Error() {
			t.Fatalf("Expecting error %v but got %v", eErr, err)
		}
	}
}
