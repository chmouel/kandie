package kandie

import (
	"context"
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/chmouel/kandie/pkg/ui"
	corev1 "k8s.io/api/core/v1"
)

func colorPhase(phase corev1.PodPhase) string {
	var color lipgloss.Color
	switch string(phase) {
	case "Running":
		color = lipgloss.Color("#0096FF6")
	case "Pending":
		color = lipgloss.Color("220")
	case "Succeeded":
		color = lipgloss.Color("76")
	case "Failed":
		color = lipgloss.Color("196")
	default:
		color = lipgloss.Color("240")
	}
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(color).
		Render(string(phase))
}

// sort pods by creation time
type podListSort []corev1.Pod

func (pods podListSort) Len() int      { return len(pods) }
func (pods podListSort) Swap(i, j int) { pods[i], pods[j] = pods[j], pods[i] }
func (pods podListSort) Less(i, j int) bool {
	// TODO: there is something weird going on when only one running pod,
	// running should be always at the top and that not's what i see here.
	if pods[i].Status.Phase == "Running" && pods[j].Status.Phase == "Running" {
		return pods[i].Status.StartTime.Before(pods[j].Status.StartTime)
	} else if pods[i].Status.Phase == "Running" || pods[j].Status.Phase == "Running" {
		return true
	}
	return pods[j].CreationTimestamp.Before(&pods[i].CreationTimestamp)
}

func (a *App) choosePod(ctx context.Context) (string, error) {
	podList, err := a.kc.podList(ctx)
	if err != nil {
		return "", fmt.Errorf("cannot get pod listing: %w", err)
	}
	sort.Sort(podListSort(podList.Items))
	columns := []table.Column{
		{
			Title: "Name",
			Width: 50,
		},
		{
			Title: "Created",
			Width: 20,
		},
		{
			Title: "Status",
			Width: 30,
		},
	}
	rows := make([]table.Row, len(podList.Items))
	for i, pod := range podList.Items {
		rows[i] = table.Row{
			pod.GetName(),
			pod.GetCreationTimestamp().Format("2006-01-02 15:04:05"),
			colorPhase(pod.Status.Phase),
		}
	}
	selected, err := ui.NewTable(columns, rows)
	if err != nil {
		return "", err
	}
	if selected == "" {
		return "", fmt.Errorf("need a pod to be selected")
	}
	return selected, nil
}

func (a *App) doPod(ctx context.Context) error {
	if a.target == "" {
		var err error
		if a.target, err = a.choosePod(ctx); err != nil {
			return err
		}
	}
	return nil
}
