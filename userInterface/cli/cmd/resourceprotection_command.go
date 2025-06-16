package cmd

import (
	"context"
	"fmt"

	"github.com/gardener/etcd-druid/userInterface/core"
	userInterfacePkg "github.com/gardener/etcd-druid/userInterface/pkg"
	"github.com/spf13/cobra"
)

var addComponentProtectionCmd = &cobra.Command{
	Use:   "add-component-protection <etcd-resource-name>",
	Short: "Adds resource protection to all managed components for a given etcd cluster",
	Long: `Adds resource protection to all managed components for a given etcd cluster.

NOTE: This will only have effect if resource protection webhook has been enabled when deploying etcd-druid.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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

		service := core.NewEtcdProtectionService(clientSet.DruidV1alpha1())
		if err := service.AddDisableProtectionAnnotation(context.TODO(), etcdResourceName, namespace, allNamespaces); err != nil {
			return err
		}
		fmt.Printf("Added protection annotation to Etcd '%s'\n", etcdResourceName)
		return nil
	},
}

var removeComponentProtectionCmd = &cobra.Command{
	Use:   "remove-component-protection <etcd-resource-name>",
	Short: "Removes resource protection for all managed components for a given etcd cluster",
	Long: `Removes resource protection for all managed components for a given etcd cluster.

NOTE: This will only have effect if resource protection webhook has been enabled when deploying etcd-druid.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		clientSet, err := userInterfacePkg.CreateTypedClientSet(configFlags)
		if err != nil {
			return fmt.Errorf("unable to create etcd typed client: %w", err)
		}
		etcdResourceName := args[0]
		namespace, _, err := configFlags.ToRawKubeConfigLoader().Namespace()
		if err != nil {
			return fmt.Errorf("failed to get namespace: %w", err)
		}
		service := core.NewEtcdProtectionService(clientSet.DruidV1alpha1())
		if err := service.RemoveDisableProtectionAnnotation(context.TODO(), etcdResourceName, namespace, allNamespaces); err != nil {
			return err
		}
		fmt.Printf("Removed protection annotation from Etcd '%s'\n", etcdResourceName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addComponentProtectionCmd)
	rootCmd.AddCommand(removeComponentProtectionCmd)
}
