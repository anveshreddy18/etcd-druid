package tui

import (
	"fmt"
	os "os"

	tea "github.com/charmbracelet/bubbletea"
	client "github.com/gardener/etcd-druid/druidctl/client"
	"github.com/gardener/etcd-druid/druidctl/pkg"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func RunTUI(configFlags *genericclioptions.ConfigFlags) {
	typedClientset, err := client.CreateTypedClientSet(configFlags)
	if err != nil {
		fmt.Println("Error creating k8s client:", err)
		os.Exit(1)
	}
	genericClientSet, err := pkg.CreateGenericClientSet(configFlags)
	if err != nil {
		fmt.Println("Error creating generic k8s client:", err)
		os.Exit(1)
	}
	m := NewModel(typedClientset, genericClientSet)
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running TUI:", err)
		os.Exit(1)
	}
}
