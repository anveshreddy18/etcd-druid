package tui

import (
	"fmt"
	os "os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gardener/etcd-druid/userInterface/pkg"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func RunTUI(configFlags *genericclioptions.ConfigFlags) {
	clientset, err := pkg.CreateTypedClientSet(configFlags)
	if err != nil {
		fmt.Println("Error creating k8s client:", err)
		os.Exit(1)
	}
	m := NewModel(clientset)
	// @anveshreddy18 -- make this a full screen at the end.
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running TUI:", err)
		os.Exit(1)
	}
}
