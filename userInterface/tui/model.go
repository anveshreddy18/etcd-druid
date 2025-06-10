package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	clientset "github.com/gardener/etcd-druid/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
)

type model struct {
	state            screenState
	etcdList         list.Model
	typedClientset   *clientset.Clientset
	genericClientSet kubernetes.Interface
	loading          bool
	err              error
	width            int
	height           int

	selectedEtcd  etcdListItem
	podList       list.Model
	selectedPod   Pod
	content       string
	containerList list.Model
	containers    []string
	viewport      viewport.Model
}

func NewModel(typedClientset *clientset.Clientset, genericClientSet kubernetes.Interface) model {
	etcdList := list.New([]list.Item{}, list.NewDefaultDelegate(), defaultWidth, defaultHeight) // set a default size
	podList := list.New([]list.Item{}, list.NewDefaultDelegate(), defaultWidth, defaultHeight)
	containerList := list.New([]list.Item{}, list.NewDefaultDelegate(), defaultWidth, defaultHeight)
	vp := viewport.New(defaultWidth*2, defaultHeight*2)
	etcdList.Title = "Etcd Clusters"
	return model{
		state:            ScreenEtcdList,
		etcdList:         etcdList,
		podList:          podList,
		containerList:    containerList,
		typedClientset:   typedClientset,
		genericClientSet: genericClientSet,
		loading:          true,
		width:            defaultWidth,
		height:           defaultHeight,
		viewport:         vp,
	}
}

func (m model) Init() tea.Cmd {
	return m.fetchEtcdsCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.etcdList.SetSize(msg.Width, msg.Height-4) // leave space for title
		m.podList.SetSize(msg.Width, msg.Height-4)
		m.containerList.SetSize(msg.Width, msg.Height-4)
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 4
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.loading {
			return m, nil
		}
		switch m.state {
		case ScreenEtcdList:
			m.etcdList, cmd = m.etcdList.Update(msg)
			cmds = append(cmds, cmd)

			switch msg.String() {
			case "enter":
				if len(m.etcdList.Items()) > 0 {
					item := m.etcdList.SelectedItem().(etcdListItem)
					m.selectedEtcd = item
					m.state = ScreenPodList
					return m, m.fetchPodsCmd(item)
				}
			case "d":
				// Add Disable protection Annotation
				if len(m.etcdList.Items()) > 0 {
					item := m.etcdList.SelectedItem().(etcdListItem)
					m.selectedEtcd = item
					return m, m.addDisableProtectionAnnotationCmd(item)
				}
			case "p":
				// Remove disable protection (i.e Protect)
				if len(m.etcdList.Items()) > 0 {
					item := m.etcdList.SelectedItem().(etcdListItem)
					m.selectedEtcd = item
					return m, m.removeProtectionAnnotationCmd(item)
				}
			case "q":
				return m, tea.Quit
			}
		case ScreenPodList:
			m.podList, cmd = m.podList.Update(msg)
			cmds = append(cmds, cmd)

			switch msg.String() {
			case "q", "esc":
				m.state = ScreenEtcdList
				// m.selectedEtcd = etcdListItem{}
				return m, nil
			case "d":
				if len(m.podList.Items()) > 0 {
					m.selectedPod = m.podList.SelectedItem().(Pod)
					m.state = ScreenPodDescribe
					return m, m.fetchDescribeCmd(m.selectedPod)
				}
			case "y":
				if len(m.podList.Items()) > 0 {
					m.selectedPod = m.podList.SelectedItem().(Pod)
					m.state = ScreenPodYAML
					return m, m.fetchYAMLCmd(m.selectedPod)
				}
			case "l", "enter":
				if len(m.podList.Items()) > 0 {
					m.selectedPod = m.podList.SelectedItem().(Pod)
					m.state = ScreenPodContainerSelect
					return m, m.fetchContainersCmd(m.selectedPod)
				}
			}
		case ScreenPodDescribe, ScreenPodYAML:
			switch msg.String() {
			case "q", "esc":
				m.state = ScreenPodList
				m.content = ""
				return m, nil
			default:
				m.viewport, cmd = m.viewport.Update(msg)
				cmds = append(cmds, cmd)
			}
		case ScreenPodLogs:
			switch msg.String() {
			case "q", "esc":
				m.state = ScreenPodContainerSelect
				m.content = ""
				return m, nil
			default:
				m.viewport, cmd = m.viewport.Update(msg)
				cmds = append(cmds, cmd)
			}
		case ScreenPodContainerSelect:
			m.containerList, cmd = m.containerList.Update(msg)
			cmds = append(cmds, cmd)

			switch msg.String() {
			case "q", "esc":
				m.state = ScreenPodList
				return m, nil
			case "enter":
				if len(m.containers) > 0 {
					container := m.containers[m.containerList.Index()]
					m.state = ScreenPodLogs
					return m, m.fetchLogsCmd(m.selectedPod, container)
				}
			}
		}
	case disableProtectionAnnotationAddedMsg:
		//
	case disableProtectionAnnotationRemovedMsg:
		//
	case errMsg:
		m.err = msg
		m.loading = false
		return m, nil
	case etcdsLoadedMsg:
		m.etcdList.SetItems(msg)
		m.etcdList.Title = "Etcds objects across Namespaces"
		m.loading = false
		return m, nil
	case podsLoadedMsg:
		items := make([]list.Item, len(msg))
		for i, pod := range msg {
			items[i] = pod
		}
		m.podList.SetItems(items)
		m.podList.Title = fmt.Sprintf("Pods for Etcd: %s/%s", m.selectedEtcd.Namespace, m.selectedEtcd.Name)
		return m, nil
	case describeLoadedMsg, logsLoadedMsg, yamlLoadedMsg:
		m.content = msg.(interface{ Content() string }).Content()
		fmt.Println("Anvesh:: content is : ", m.content)
		m.viewport.SetContent(m.content)
		return m, nil
	case containersLoadedMsg:
		m.containers = msg.containers
		items := make([]list.Item, len(msg.containers))
		for i, c := range msg.containers {
			items[i] = listItemString(c)
		}
		m.containerList.SetItems(items)
		m.containerList.Title = fmt.Sprintf("Containers for Pod: %s", m.selectedPod.Name)
		return m, nil
	}
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress 'q' to quit.", m.err)
	}
	switch m.state {
	case ScreenEtcdList:
		header := "Etcd Clusters (press Enter to view pods)"
		help := "Enter: select Etcd • d(vulnerable): add disable protection annotation • p(protect): remove disable protection annotation • q: quit"
		return fmt.Sprintf("%s\n%s\n%s", header, m.etcdList.View(), help)
	case ScreenPodList:
		header := fmt.Sprintf("Pods for Etcd: %s/%s", m.selectedEtcd.Namespace, m.selectedEtcd.Name)
		help := "d: describe pod • y: yaml • l/Enter: select container for logs • q: back"
		return fmt.Sprintf("%s\n%s\n%s", header, m.podList.View(), help)
	case ScreenPodDescribe:
		header := fmt.Sprintf("Describe: %s", m.selectedPod.Name)
		help := "esc/q: back • ↑/↓: scroll"
		return fmt.Sprintf("%s\n%s\n%s", header, m.viewport.View(), help)
	case ScreenPodLogs:
		header := fmt.Sprintf("Logs: %s", m.selectedPod.Name)
		help := "esc/q: back • ↑/↓: scroll"
		return fmt.Sprintf("%s\n%s\n%s", header, m.viewport.View(), help)
	case ScreenPodYAML:
		header := fmt.Sprintf("YAML: %s", m.selectedPod.Name)
		help := "esc/q: back • ↑/↓: scroll"
		return fmt.Sprintf("%s\n%s\n%s", header, m.viewport.View(), help)
	case ScreenPodContainerSelect:
		header := fmt.Sprintf("Select Container: %s", m.selectedPod.Name)
		help := "Enter: show logs for container • esc/q: back"
		return fmt.Sprintf("%s\n%s\n%s", header, m.containerList.View(), help)
	}
	return ""
}
