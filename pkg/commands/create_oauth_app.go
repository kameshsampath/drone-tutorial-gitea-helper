package commands

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//OAuthAppOptions to hold the options to create oauth application
type OAuthAppOptions struct {
	oAuthAppName        string
	appRedirectURL      string
	giteaAdminPassword  string
	giteaAdminUser      string
	giteaURL            string
	addKubernetesSecret bool
	namespace           string
	kubeconfig          string
}

// OAuthAppOptions implements Interface
var _ Command = (*OAuthAppOptions)(nil)

var oAuthAppCommandExample = fmt.Sprintf(`
  # Create oAuthApp with defaults
  %[1]s oauthapp --app-name my-app 
  # Create oAuthApp with app and host
  %[1]s oauthapp -a my-app -h http://example.com
  # Create oAuthApp with app,host,gitea url, admin and password
  %[1]s oauthapp -a my-app -h http://example.com -g https://try.gitea.com -u myAdmin -p myAdmin123
  # Create oAuthApp and store the client id and secret in kubernetes secret
  %[1]s oauthapp --app-name my-app  -s -n my-namesapce
`, ExamplePrefix())

//NewCreateOAuthAppCommand instantiates the new instance of the StartCommand
func NewCreateOAuthAppCommand() *cobra.Command {
	oAuthOpts := &OAuthAppOptions{}

	oAuthCmd := &cobra.Command{
		Use:     "oauthapp",
		Short:   "Create an Gitea OAuthApp",
		Example: oAuthAppCommandExample,
		RunE:    oAuthOpts.Execute,
		PreRunE: oAuthOpts.Validate,
	}

	oAuthOpts.AddFlags(oAuthCmd)

	return oAuthCmd
}

// AddFlags implements Command
func (opts *OAuthAppOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&opts.oAuthAppName, "app-name", "a", "", "The Gitea oAuth Application Name")
	if err := cmd.MarkFlagRequired("app-name"); err != nil {
		log.Fatalf("Error marking flag 'app-name' as required %v", err)
	}
	cmd.Flags().StringVarP(&opts.appRedirectURL, "app-redirect-url", "r", "http://drone-127.0.0.1.sslip.io:30980", "The Gitea oAuth Application Redirect URL")
	cmd.Flags().StringVarP(&opts.giteaAdminUser, "gitea-admin-user", "u", "demo", "The Gitea admin username")
	cmd.Flags().StringVarP(&opts.giteaAdminPassword, "gitea-admin-password", "p", "demo@123", "The Gitea admin user password")
	cmd.Flags().StringVarP(&opts.giteaURL, "gitea-url", "g", "http://gitea-127.0.0.1.sslip.io:30950", "The Gitea URL")
	cmd.Flags().BoolVarP(&opts.addKubernetesSecret, "add-k8s-secret", "s", false, "Create a Kubernetes secret with oAuth application name, to hold the client id and client secret of the oAuth application")
	cmd.Flags().StringVarP(&opts.namespace, "k8s-namespace", "n", "", "The namespace where to create the kubernetes secret for the oAuth application")
	cmd.Flags().StringVarP(&opts.kubeconfig, "kubeconfig", "k", "", "The kubeconfig file to use")
}

// Execute implements Command
func (opts *OAuthAppOptions) Execute(cmd *cobra.Command, args []string) error {
	wopts := &WorkshopOptions{
		GiteaURL:           opts.giteaURL,
		GiteaAdminUser:     opts.giteaAdminUser,
		GiteaAdminPassword: opts.giteaAdminPassword,
	}
	c, err := wopts.newGiteaClient()
	if err != nil {
		return err
	}

	_, err = opts.createOAuthApp(c)

	if err != nil {
		return err
	}
	return nil
}

// Validate implements Command
func (opts *OAuthAppOptions) Validate(cmd *cobra.Command, args []string) error {
	err := viper.BindPFlags(cmd.Flags())

	if err != nil {
		return err
	}

	if opts.addKubernetesSecret = viper.GetBool("add-k8s-secret"); opts.addKubernetesSecret {
		if opts.namespace = viper.GetString("k8s-namespace"); opts.namespace == "" {
			return fmt.Errorf("require namespace to create the %s secret", opts.oAuthAppName)
		}
	}
	return nil
}
