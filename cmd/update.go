package commands

import (
	"github.com/eris-ltd/eris-cli/update"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

var Update = &cobra.Command{
	Use:     "update",
	Aliases: []string{"upgrade"},
	Short:   "Update the eris tool.",
	Long: `Fetch the latest version (master branch by default)
and re-install eris. Once eris is reinstalled, then the
eris init function will be called automatically for you
in order to update your definition files and images.

If you have made modifications to the default definition files
then you will want to make backups of those **before** upgrading
your eris installation.`,
	Run: func(cmd *cobra.Command, args []string) {
		UpdateTool(cmd, args)
	},
}

func buildUpdateCommand() {
	addUpdateFlags()
}

func addUpdateFlags() {
	Update.Flags().StringVarP(&do.Branch, "branch", "b", "master", "specify a branch to update from. requires that eris was installed via go and, git installed")
	Update.Flags().StringVarP(&do.Commit, "commit", "", "", "specify a commit to update from. requires that eris was installed via go and, git installed")
	Update.Flags().StringVarP(&do.Version, "version", "", "", "specify a branch to update from. requires that eris was installed via go and, git installed")
}

func UpdateTool(cmd *cobra.Command, args []string) {
	//arg check?
	update.UpdateEris(do)
}
