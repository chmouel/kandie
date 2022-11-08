package kandie

import (
	"context"
	_ "embed"
	"fmt"
	"sort"
	"text/tabwriter"
	"text/template"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	uitable "github.com/chmouel/kandie/pkg/ui/table"
	"github.com/juju/ansiterm"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:embed templates/describe_pod.tmpl
var describePodTmpl string

func colorPhase(phase string) string {
	var color string
	switch phase {
	case "Running":
		color = "#0096FF6"
	case "Pending":
		color = "220"
	case "Succeeded":
		color = "76"
	case "Failed":
		color = "196"
	default:
		color = "240"
	}
	return ColorIt(color, phase)
}

func ColorIt(color, str string) string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(color)).
		Render(string(str))
}

func ColorItBackground(fgcolor, bgcolor, str string) string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(fgcolor)).
		Background(lipgloss.Color(bgcolor)).
		Render(string(str))
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
			colorPhase(string(pod.Status.Phase)),
		}
	}
	selected, err := uitable.NewTable(columns, rows)
	if err != nil {
		return "", err
	}
	if selected == "" {
		return "", fmt.Errorf("need a pod to be selected")
	}
	return selected, nil
}

func getPodCounts(pod corev1.Pod) ([]corev1.ContainerStatus, []corev1.ContainerStatus, []corev1.ContainerStatus, []corev1.ContainerStatus) {
	running := []corev1.ContainerStatus{}
	good := []corev1.ContainerStatus{}
	failed := []corev1.ContainerStatus{}
	waiting := []corev1.ContainerStatus{}
	for _, s := range append(pod.Status.ContainerStatuses, pod.Status.InitContainerStatuses...) {
		if s.State.Waiting != nil && s.State.Waiting.Reason == "ImagePullBackOff" {
			failed = append(failed, s)
			continue
		}
		if s.State.Waiting != nil {
			waiting = append(waiting, s)
			continue
		}

		if s.State.Running != nil {
			running = append(failed, s)
			continue
		}

		if s.State.Terminated != nil && s.State.Terminated.ExitCode != 0 {
			failed = append(failed, s)
			continue
		}
		good = append(good, s)
	}
	return running, good, failed, waiting
}

func GetPodStatus(pod corev1.Pod) string {
	allr, _, allf, allp := getPodCounts(pod)
	if len(allr) > 0 {
		return colorPhase("Running")
	}

	if len(allp) > 0 {
		return colorPhase("Pending")
	}

	if len(allf) > 0 {
		return colorPhase("Failed")
	}

	return colorPhase("Succeeded")
}

func (a *App) doPod(ctx context.Context) error {
	if a.target == "" {
		var err error
		if a.target, err = a.choosePod(ctx); err != nil {
			return err
		}
	}

	pod, err := a.kc.clientset.CoreV1().Pods(a.kc.namespace).Get(ctx, a.target, v1.GetOptions{})
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"C":         ColorIt,
		"CB":        ColorItBackground,
		"PodStatus": GetPodStatus,
	}
	data := struct{ Pod *corev1.Pod }{Pod: pod}

	w := ansiterm.NewTabWriter(a.iost.Out, 0, 5, 3, ' ', tabwriter.TabIndent)
	t := template.Must(template.New("Describe Pod").Funcs(funcMap).Parse(describePodTmpl))
	if err := t.Execute(w, data); err != nil {
		return err
	}

	w.Flush()
	return nil
}
