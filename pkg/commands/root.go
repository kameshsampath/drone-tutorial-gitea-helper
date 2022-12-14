package commands

import (
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var v string

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "drone-workshopper",
		Short: "An helper to interact and configure Gitea using its REST API",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := logSetup(os.Stdout, v); err != nil {
				return err
			}
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVarP(&v, "verbose", "v", log.WarnLevel.String(), "The logging level to set")

	rootCmd.AddCommand(NewVersionCommand())
	rootCmd.AddCommand(NewWorkshopSetupCommand())

	return rootCmd
}

func logSetup(out io.Writer, level string) error {
	log.SetOutput(out)
	lvl, err := log.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(lvl)
	return nil
}
