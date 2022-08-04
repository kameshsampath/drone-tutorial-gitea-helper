package commands

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	log "github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func (opts *OAuthAppOptions) createOAuthApp(c *gitea.Client) (*gitea.Oauth2, error) {
	var oAuthApp *gitea.Oauth2

	oAuthApps, _, err := c.ListOauth2(gitea.ListOauth2Option{})

	if err != nil {
		return nil, err
	}

	var appExists = false
	for _, o := range oAuthApps {
		if o.Name == opts.oAuthAppName {
			appExists = true
			break
		}
	}

	if !appExists {
		log.Debugln("Creating new oAuth App")

		oAuthApp, _, err := c.CreateOauth2(gitea.CreateOauth2Option{
			RedirectURIs: []string{fmt.Sprintf("%s/login", opts.appRedirectURL)},
			Name:         opts.oAuthAppName})
		if err != nil {
			return nil, err
		}

		if opts.addKubernetesSecret {
			err = opts.generateKubernetesSecret(oAuthApp)

			if err != nil {
				return nil, err
			}
		}
		log.Infof("\nSuccessfully created oAuth application %s\n", opts.oAuthAppName)
		log.Debugf("\noAuth application %s ClientID:%s ClientSecret:%s\n", opts.oAuthAppName, oAuthApp.ClientID, oAuthApp.ClientSecret)
	} else {
		log.Infof("\noAuth app %s already exists, updating", opts.oAuthAppName)
		oAuthApp, _, err = c.UpdateOauth2(oAuthApp.ID,
			gitea.CreateOauth2Option{
				RedirectURIs: []string{opts.appRedirectURL},
				Name:         opts.oAuthAppName,
			})
		if err != nil {
			return nil, err
		}
	}

	return oAuthApp, nil
}

// generateKubernetesSecret generates a Kubernetes secret
// for the oAuth Application and stores the ClientID and ClientSecret in it.
// The default name of the secret is <oauth-app-name>-secret
func (opts *OAuthAppOptions) generateKubernetesSecret(o *gitea.Oauth2) error {
	var config *rest.Config
	var err error
	if opts.kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", opts.kubeconfig)
		if err != nil {
			return err
		}
		log.Debugln("Using out of Cluster Config")
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return err
		}
		log.Debugln("Using InCluster Config")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	log.Debugln("Got Client Set")

	sec, _ := randomHex(16)

	//use defaults namespace
	if opts.namespace == "" {
		opts.namespace = "default"
	}

	_, err = clientset.CoreV1().Secrets(opts.namespace).Create(context.TODO(), &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-secret", opts.oAuthAppName),
		},
		StringData: map[string]string{
			"DRONE_GITEA_CLIENT_ID":     o.ClientID,
			"DRONE_GITEA_CLIENT_SECRET": o.ClientSecret,
			"DRONE_RPC_SECRET":          sec,
		},
	}, metav1.CreateOptions{})

	if err != nil {
		return err
	}
	log.Infof("Created Kubernetes secret %s", opts.oAuthAppName)
	return nil
}

func createRepo(c *gitea.Client, githubTemplateRepo, user, repoName string) error {
	repo, _, err := c.GetRepo(user, repoName)

	if err != nil {
		//raise panic if error is not repo not found or repo exists
		if strings.TrimSpace(err.Error()) != "404 Not Found" && strings.TrimSpace(err.Error()) != "409 Conflict" {
			return err
		}
	}

	if repo == nil || repo.Name == "" {
		newR, _, err := c.MigrateRepo(gitea.MigrateRepoOption{
			CloneAddr: githubTemplateRepo,
			RepoOwner: user,
			RepoName:  repoName,
		})

		if err != nil {
			return err
		}
		log.Infof("Repo %s successfully created for user %s, you can clone via %s", newR.Name, user, repo.CloneURL)
	} else {
		log.Infof("Repo %s already exists for user %s skipping creation,you can clone via %s", repo.Name, user, repo.CloneURL)
	}

	return nil
}
