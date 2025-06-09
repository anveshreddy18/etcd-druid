package cmd

import "github.com/spf13/cobra"

var (
	waitTillReadyFlag bool
)

var reconcileCmd = &cobra.Command{
	Use:   "reconcile <etcd-resource-name> --wait-till-ready(optional flag)",
	Short: "Reconcile the mentioned etcd resource",
	Long:  `Reconcile the mentioned etcd resource. If the flag --wait-till-ready is set, then reconcile only after the Etcd CR is considered ready`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Your logic here, you can use waitTillReadyFlag
		return nil
	},
}

func init() {
	reconcileCmd.Flags().BoolVar(&waitTillReadyFlag, "wait-till-ready", false, "Wait until the Etcd resource is ready before reconciling")
	rootCmd.AddCommand(reconcileCmd)
}
