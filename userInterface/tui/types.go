package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
)

type screenState int

const (
	ScreenEtcdList screenState = iota
	ScreenPodList
	ScreenPodLogs
	ScreenPodDescribe
	ScreenPodYAML
	ScreenPodContainerSelect
)

const (
	defaultWidth  int = 40
	defaultHeight int = 10
)

type etcdListItem struct {
	Name      string
	Namespace string
}

func (e etcdListItem) Title() string       { return e.Name }
func (e etcdListItem) Description() string { return fmt.Sprintf("Namespace: %s", e.Namespace) }
func (e etcdListItem) FilterValue() string { return e.Name }

type Pod struct {
	Name      string
	Namespace string
	Status    string
	Ready     string
	Age       string
	Node      string
}

func (p Pod) FilterValue() string { return p.Name }
func (p Pod) Title() string       { return p.Name }
func (p Pod) Description() string {
	return fmt.Sprintf("Status: %s | Ready: %s | Node: %s | Age: %s", p.Status, p.Ready, p.Node, p.Age)
}

type listItemString string

func (s listItemString) Title() string       { return string(s) }
func (s listItemString) Description() string { return "" }
func (s listItemString) FilterValue() string { return string(s) }

// Message types for async commands
type podsLoadedMsg []Pod
type describeLoadedMsg struct{ content string }
type logsLoadedMsg struct{ content string }
type yamlLoadedMsg struct{ content string }
type containersLoadedMsg struct{ containers []string }
type etcdsLoadedMsg []list.Item
type errMsg struct{ error }
type disableProtectionAnnotationAddedMsg struct{}
type disableProtectionAnnotationRemovedMsg struct{}

func (m describeLoadedMsg) Content() string { return m.content }
func (m logsLoadedMsg) Content() string     { return m.content }
func (m yamlLoadedMsg) Content() string     { return m.content }
