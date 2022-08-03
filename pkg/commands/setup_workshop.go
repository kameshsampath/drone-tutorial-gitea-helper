package commands

import (
	"fmt"
	"io/ioutil"

	"code.gitea.io/sdk/gitea"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	yamlv2 "gopkg.in/yaml.v2"
)

//WorkshopSetupOptions the configuration data for workshop
type WorkshopSetupOptions struct {
	configFile string
	kubeconfig string
}

//WorkshopOptions the configuration data for workshop
type WorkshopOptions struct {
	GiteaAdminPassword string    `yaml:"giteaAdminUserPassword,omitempty"`
	GiteaAdminUser     string    `yaml:"giteaAdminUserName,omitempty"`
	GiteaURL           string    `yaml:"giteaURL,omitempty"`
	GiteaUsers         GiteaUser `yaml:"users"`
}

//GiteaUser is a Gitea user
type GiteaUser struct {
	From                int      `yaml:"from"`
	To                  int      `yaml:"to"`
	AddKubernetesSecret bool     `yaml:"addKubernetesSecret"`
	Namespace           string   `yaml:"namespace"`
	OAuthAppName        string   `yaml:"oAuthAppName"`
	OAuthRedirectURI    string   `yaml:"oAuthRedirectURI"`
	SecretNamespace     string   `yaml:"secretNamespace"`
	Repos               []string `yaml:"repos"`
}

// WorkshopOptions implements Interface
var _ Command = (*WorkshopSetupOptions)(nil)

var workshopCommandExample = fmt.Sprintf(`
  # Create oAuthApp with defaults
  %[1]s setup-workshop --workshop-file my-app
  # Create oAuthApp and store the client id and secret in kubernetes secret
  %[1]s setup-workshop --app-name my-app  -k ~/.kube/config
`, ExamplePrefix())

//NewWorkshopSetupCommand instantiates the new instance of the NewWorkshopSetupCommand
func NewWorkshopSetupCommand() *cobra.Command {
	workshopSetupOpts := &WorkshopSetupOptions{}

	workshopSetupCmd := &cobra.Command{
		Use:     "setup-workshop",
		Short:   "Setup Workshop",
		Example: workshopCommandExample,
		RunE:    workshopSetupOpts.Execute,
		PreRunE: workshopSetupOpts.Validate,
	}

	workshopSetupOpts.AddFlags(workshopSetupCmd)

	return workshopSetupCmd
}

// AddFlags implements Command
func (opts *WorkshopSetupOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&opts.configFile, "workshop-file", "f", "", "The workshop configuration file")
	if err := cmd.MarkFlagRequired("workshop-file"); err != nil {
		log.Fatalf("Error marking flag 'workshop-file' as required %v", err)
	}
	cmd.Flags().StringVarP(&opts.kubeconfig, "kubeconfig", "k", "", "The kubeconfig file to use")
}

// Execute implements Command
func (opts *WorkshopSetupOptions) Execute(cmd *cobra.Command, args []string) error {
	var workshopOpts WorkshopOptions
	b, err := ioutil.ReadFile(opts.configFile)
	if err != nil {
		return err
	}
	err = yamlv2.Unmarshal(b, &workshopOpts)
	if err != nil {
		return err
	}

	log.Debugf("%#v", workshopOpts)

	_, err = workshopOpts.createUsers(opts.kubeconfig)

	if err != nil {
		return err
	}

	return nil
}

func (opts *WorkshopOptions) createUsers(kubeconfig string) ([]*gitea.User, error) {
	log.Debugln("Creating users")
	var err error
	giteaUsers := opts.GiteaUsers

	gusers := make([]*gitea.User, giteaUsers.To)

	c, err := opts.newGiteaClient()

	if err != nil {
		return nil, err
	}

	for i := giteaUsers.From; i <= giteaUsers.To; i++ {
		cp := false
		userName := fmt.Sprintf("user-%02d", i)
		userEmail := fmt.Sprintf("user-%02d@example.com", i)
		userPassword := fmt.Sprintf("user-%02d@123", i)

		if u, _, err := c.GetUserInfo(userName); u != nil && err == nil {
			log.Infof("User %s already exists", u.UserName)
			continue
		}

		if err != nil {
			return nil, err
		}

		uOpt := gitea.CreateUserOption{
			Username:           userName,
			Email:              userEmail,
			Password:           userPassword,
			MustChangePassword: &cp,
			SendNotify:         false,
		}

		u, _, err := c.AdminCreateUser(uOpt)

		if err != nil {
			return nil, err
		}
		log.Infof("Created user with username %s", u.UserName)
		gusers = append(gusers, u)

		//Create oAuth2 App
		c.SetSudo(u.UserName)

		oauthOpts := OAuthAppOptions{
			oAuthAppName:        fmt.Sprintf("%s-user-%02d", giteaUsers.OAuthAppName, i),
			appRedirectURL:      fmt.Sprintf("%s/login", giteaUsers.OAuthRedirectURI),
			addKubernetesSecret: giteaUsers.AddKubernetesSecret,
			namespace:           giteaUsers.SecretNamespace,
			kubeconfig:          kubeconfig,
		}

		_, err = oauthOpts.createOAuthApp(c)
		if err != nil {
			return nil, err
		}

		for _, repoURL := range giteaUsers.Repos {
			repoName, err := repoNameFromURL(repoURL)
			if err != nil {
				return nil, err
			}
			if err := createRepo(c, repoURL, u.UserName, repoName); err != nil {
				return nil, err
			}
		}

		//Set it back to admin
		c.SetSudo(opts.GiteaAdminUser)
	}

	return gusers, nil
}

// Validate implements Command
func (opts *WorkshopSetupOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}
