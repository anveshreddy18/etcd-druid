package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// fetchEtcdPods retrieves pods managed by the StatefulSet that corresponds to our Etcd resource
func (m *model) fetchEtcdPods(etcdName, namespace string) ([]Pod, error) {
	labelSelector := fmt.Sprintf("app.kubernetes.io/name=%s", etcdName)
	podList, err := m.genericClientSet.CoreV1().Pods(namespace).List(
		context.Background(),
		metav1.ListOptions{LabelSelector: labelSelector},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list etcd pods: %w", err)
	}
	var pods []Pod
	for _, pod := range podList.Items {
		age := time.Since(pod.CreationTimestamp.Time).Truncate(time.Second)
		readyCount := 0
		totalCount := len(pod.Status.ContainerStatuses)
		for _, status := range pod.Status.ContainerStatuses {
			if status.Ready {
				readyCount++
			}
		}
		pods = append(pods, Pod{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    string(pod.Status.Phase),
			Ready:     fmt.Sprintf("%d/%d", readyCount, totalCount),
			Age:       age.String(),
			Node:      pod.Spec.NodeName,
		})
	}
	return pods, nil
}

// fetchPodContainers retrieves the list of containers for a given pod
func (m *model) fetchPodContainers(podName, namespace string) ([]string, error) {
	pod, err := m.genericClientSet.CoreV1().Pods(namespace).Get(
		context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %s: %w", podName, err)
	}
	var containers []string
	for _, c := range pod.Spec.Containers {
		containers = append(containers, c.Name)
	}
	return containers, nil
}

// fetchPodYAML retrieves the YAML configuration for a given pod
func (m *model) fetchPodYAML(podName, namespace string) (string, error) {
	pod, err := m.genericClientSet.CoreV1().Pods(namespace).Get(
		context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get pod %s: %w", podName, err)
	}
	b, err := yaml.Marshal(pod)
	if err != nil {
		return "", fmt.Errorf("failed to marshal pod to yaml: %w", err)
	}
	return string(b), nil
}

// getPodLogs retrieves logs for the selected pod and container
func (m *model) getPodLogs(podName, namespace, container string) (string, error) {
	tailLines := int64(100)
	req := m.genericClientSet.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		TailLines: &tailLines,
		Container: container,
	})
	logs, err := req.Stream(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get logs for pod %s (container %s): %w", podName, container, err)
	}
	defer logs.Close()
	buf := make([]byte, 2048)
	var result strings.Builder
	for {
		n, err := logs.Read(buf)
		if n > 0 {
			result.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
	return result.String(), nil
}

// describePod gets detailed information about a pod
func (m *model) describePod(podName, namespace string) (string, error) {
	fmt.Println("Anvesh:: Its coming to the describePod with podname and namespace as :", podName, namespace)
	pod, err := m.genericClientSet.CoreV1().Pods(namespace).Get(
		context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to describe pod %s: %w", podName, err)
	}
	var desc strings.Builder
	desc.WriteString(fmt.Sprintf("Name: %s\n", pod.Name))
	desc.WriteString(fmt.Sprintf("Namespace: %s\n", pod.Namespace))
	desc.WriteString(fmt.Sprintf("Node: %s\n", pod.Spec.NodeName))
	desc.WriteString(fmt.Sprintf("Status: %s\n", pod.Status.Phase))
	desc.WriteString(fmt.Sprintf("IP: %s\n", pod.Status.PodIP))
	desc.WriteString(fmt.Sprintf("Created: %s\n", pod.CreationTimestamp.Time.Format(time.RFC3339)))
	desc.WriteString("\nContainers:\n")
	for _, container := range pod.Spec.Containers {
		desc.WriteString(fmt.Sprintf("  %s: %s\n", container.Name, container.Image))
	}
	desc.WriteString("\nConditions:\n")
	for _, condition := range pod.Status.Conditions {
		desc.WriteString(fmt.Sprintf("  %s: %s\n", condition.Type, condition.Status))
	}
	return desc.String(), nil
}
