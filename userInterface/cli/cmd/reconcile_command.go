package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/gardener/etcd-druid/userInterface/core"
	userInterfacePkg "github.com/gardener/etcd-druid/userInterface/pkg"
	"github.com/spf13/cobra"
)

var (
	waitTillReadyFlag bool
	timeoutFlag       time.Duration
)

var reconcileCmd = &cobra.Command{
	Use:   "reconcile <etcd-resource-name> --wait-till-ready(optional flag)",
	Short: "Reconcile the mentioned etcd resource",
	Long:  `Reconcile the mentioned etcd resource. If the flag --wait-till-ready is set, then reconcile only after the Etcd CR is considered ready`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Your logic here, you can use waitTillReadyFlag
		var (
			etcdResourceName string
			namespace        string
			err              error
		)

		clientSet, err := userInterfacePkg.CreateTypedClientSet(configFlags)
		if err != nil {
			return fmt.Errorf("unable to create etcd typed client: %w", err)
		}

		if !allNamespaces {
			etcdResourceName = args[0]
			if namespace, _, err = configFlags.ToRawKubeConfigLoader().Namespace(); err != nil {
				return fmt.Errorf("failed to get namespace: %w", err)
			}
		}

		service := core.NewEtcdReconciliationService(clientSet.DruidV1alpha1(), waitTillReadyFlag, timeoutFlag)
		if err := service.ReconcileEtcd(context.TODO(), etcdResourceName, namespace, allNamespaces); err != nil {
			return err
		}
		fmt.Println("Reconciliation Done for Etcd: ns: ", namespace, " name: ", etcdResourceName)
		return nil
	},
}

func init() {
	reconcileCmd.Flags().BoolVarP(&waitTillReadyFlag, "wait-till-ready", "w", false, "Wait until the Etcd resource is ready before reconciling")
	reconcileCmd.Flags().DurationVarP(&timeoutFlag, "timeout", "t", 5*time.Minute, "Timeout for the reconciliation process")
	rootCmd.AddCommand(reconcileCmd)
}
