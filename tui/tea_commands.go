package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gardener/etcd-druid/druidctl/cli/types"
	client "github.com/gardener/etcd-druid/druidctl/client"
	core "github.com/gardener/etcd-druid/druidctl/internal"
	"github.com/gardener/etcd-druid/druidctl/pkg"
)

// Async commands for Etcds, pods, describe, logs, yaml, containers
func (m model) fetchEtcdsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		items, err := pkg.ListAllEtcds(ctx, m.typedClientset)
		if err != nil {
			return errMsg{err}
		}
		var listItems []list.Item
		for _, e := range items {
			listItems = append(listItems, etcdListItem{Name: e.Name, Namespace: e.Namespace})
		}
		return etcdsLoadedMsg(listItems)
	}
}

func (m model) fetchPodsCmd(etcd etcdListItem) tea.Cmd {
	return func() tea.Msg {
		pods, err := m.fetchEtcdPods(etcd.Name, etcd.Namespace)
		if err != nil {
			return errMsg{err}
		}
		return podsLoadedMsg(pods)
	}
}
func (m model) fetchDescribeCmd(pod Pod) tea.Cmd {
	fmt.Println("It came to the fetchDescribeCmd function")
	return func() tea.Msg {
		content, err := m.describePod(pod.Name, pod.Namespace)
		if err != nil {
			return errMsg{err}
		}
		return describeLoadedMsg{content}
	}
}
func (m model) fetchLogsCmd(pod Pod, container string) tea.Cmd {
	return func() tea.Msg {
		content, err := m.getPodLogs(pod.Name, pod.Namespace, container)
		if err != nil {
			return errMsg{err}
		}
		return logsLoadedMsg{content}
	}
}
func (m model) fetchYAMLCmd(pod Pod) tea.Cmd {
	return func() tea.Msg {
		content, err := m.fetchPodYAML(pod.Name, pod.Namespace)
		if err != nil {
			return errMsg{err}
		}
		return yamlLoadedMsg{content}
	}
}
func (m model) fetchContainersCmd(pod Pod) tea.Cmd {
	return func() tea.Msg {
		containers, err := m.fetchPodContainers(pod.Name, pod.Namespace)
		if err != nil {
			return errMsg{err}
		}
		return containersLoadedMsg{containers}
	}
}

func (m model) addDisableProtectionAnnotationCmd(etcdItem etcdListItem) tea.Cmd {
	return func() tea.Msg {
		client := client.NewEtcdClient(m.typedClientset.DruidV1alpha1())
		cmdCtx := types.CommandContext{
			ResourceName: etcdItem.Name,
			Namespace:    etcdItem.Namespace,
		}
		resourceProtectionCtx := types.NewResourceProtectionCommandContext(&cmdCtx)
		resourceProtectionCtx.EtcdClient = client
		if err := core.AddDisableProtectionAnnotation(context.TODO(), resourceProtectionCtx); err != nil {
			return errMsg{err}
		}
		return disableProtectionAnnotationAddedMsg{}
	}
}

func (m model) removeProtectionAnnotationCmd(etcdItem etcdListItem) tea.Cmd {
	return func() tea.Msg {
		client := client.NewEtcdClient(m.typedClientset.DruidV1alpha1())
		cmdCtx := types.CommandContext{
			ResourceName: etcdItem.Name,
			Namespace:    etcdItem.Namespace,
		}
		resourceProtectionCtx := types.NewResourceProtectionCommandContext(&cmdCtx)
		resourceProtectionCtx.EtcdClient = client
		if err := core.RemoveDisableProtectionAnnotation(context.TODO(), resourceProtectionCtx); err != nil {
			return errMsg{err}
		}
		return disableProtectionAnnotationRemovedMsg{}
	}
}
