package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	clientset "github.com/gardener/etcd-druid/client/clientset/versioned"
	"github.com/gardener/etcd-druid/userInterface/pkg"
)

type screenState int

const (
	ScreenEtcdList screenState = iota
)

type etcdListItem struct {
	Name      string
	Namespace string
}

func (e etcdListItem) Title() string       { return e.Name }
func (e etcdListItem) Description() string { return fmt.Sprintf("Namespace: %s", e.Namespace) }
func (e etcdListItem) FilterValue() string { return e.Name }

type model struct {
	state   screenState
	list    list.Model
	client  *clientset.Clientset
	loading bool
	err     error
	width   int
	height  int
}

func NewModel(client *clientset.Clientset) model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 40, 10) // set a default size
	l.Title = "Etcd Clusters"
	return model{
		state:   ScreenEtcdList,
		list:    l,
		client:  client,
		loading: true,
		width:   40,
		height:  10,
	}
}

func (m model) Init() tea.Cmd {
	return m.fetchEtcdsCmd()
}

func (m model) fetchEtcdsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		items, err := pkg.ListAllEtcds(ctx, m.client)
		if err != nil {
			return errMsg{err}
		}
		var listItems []list.Item
		for _, e := range items {
			listItems = append(listItems, etcdListItem{Name: e.Name, Namespace: e.Namespace})
		}
		return etcdListMsg(listItems)
	}
}

type etcdListMsg []list.Item

type errMsg struct{ error }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.list.SetSize(msg.Width, msg.Height-2) // leave space for title
		return m, nil
	case etcdListMsg:
		m.list.SetItems(msg)
		m.loading = false
		return m, nil
	case errMsg:
		m.err = msg
		m.loading = false
		return m, nil
	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	if m.loading {
		return "Loading Etcd resources..."
	}
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}
	return m.list.View()
}
