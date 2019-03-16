package cli

import (
	"fmt"

	"github.com/apex/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:              "version",
	Short:            "Print the version number and exit",
	PersistentPreRun: noValidation,
	Args:             cobra.NoArgs,
	Run:              showVersion,
}

func noValidation(cmd *cobra.Command, args []string) {}

func showVersion(cmd *cobra.Command, args []string) {
	log.WithField("version", AppVersion()).
		Info(fmt.Sprintf("%v version", AppName()))
	log.WithField("commit", AppCommit()).
		Info("Git SCM commit hash")
	log.WithField("date", AppDate()).
		Info("Build date")
}
