package commands

import (
	"github.com/Shopify/kubeaudit/auditors/label"
	"github.com/spf13/cobra"
)

var labelConfig label.Config

const labelFlagName = "label"

var labelCmd = &cobra.Command{
	Use:   "label",
	Short: "Audit deployments not using a specified label",
	Long: `This command audits a deployment against a given label.

An ERROR result is generated when a deployment does not match the label

An INFO result is generated when a deployment has a matching label.

This command is also a root command, check 'kubeaudit label --help'.

Example usage:
kubeaudit label --label project
kubeaudit image -l project`,
	Run: func(cmd *cobra.Command, args []string) {
		runAudit(label.New(labelConfig))(cmd, args)
	},
}

func setLabelFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&labelConfig.Label, labelFlagName, "l", "", "Label to check against")
}

func init() {
	RootCmd.AddCommand(labelCmd)
	setLabelFlags(labelCmd)
}
