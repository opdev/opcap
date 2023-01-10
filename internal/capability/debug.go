package capability

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/opdev/opcap/internal/logger"
	"github.com/opdev/opcap/internal/report"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CollectDebugData(ctx context.Context, options auditOptions, reportName string) error {
	c, err := k8sClientset()
	if err != nil {
		return fmt.Errorf("couldn't get clientset for operator install debug report: %s", err.Error())
	}

	// get CSV events
	events := []report.Event{}
	if options.csv != nil {
		EventList, err := EventsByNameAndKind(ctx, c, options.csv.ObjectMeta.Name, "ClusterServiceVersion", options.namespace)
		if err != nil {
			return fmt.Errorf("couldn't get eventList for CSV: %s", err.Error())
		}

		if len(EventList.Items) > 0 {
			for _, event := range EventList.Items {
				events = append(events, report.Event{
					InvolvedObjName:   event.InvolvedObject.Name,
					InvolvedObjkind:   event.InvolvedObject.Kind,
					CreationTimestamp: event.CreationTimestamp,
					Message:           strings.Replace(event.Message, "\"", "", -1),
					Reason:            event.Reason,
				})
			}
		}
	}
	// get pods status and events
	pods, err := OperatorPods(ctx, c, options.namespace)
	if err != nil {
		logger.Infow("couldn't list pods for debug report: %s", err.Error())
	}

	podEvents := []report.Event{}
	podLogs := []report.PodLog{}
	if len(pods.Items) > 0 {
		for _, pod := range pods.Items {

			EventList, err := EventsByNameAndKind(ctx, c, pod.ObjectMeta.Name, "Pod", options.namespace)
			if err != nil {
				return fmt.Errorf("couldn't get eventList for CSV: %s", err.Error())
			}
			if len(EventList.Items) > 0 {
				for _, event := range EventList.Items {
					podEvents = append(podEvents, report.Event{
						InvolvedObjName:   event.InvolvedObject.Name,
						InvolvedObjkind:   event.InvolvedObject.Kind,
						CreationTimestamp: event.CreationTimestamp,
						Message:           strings.Replace(event.Message, "\"", "", -1),
						Reason:            event.Reason,
					})
				}
			}
			for _, container := range pod.Spec.Containers {
				log, err := Logs(ctx, c, pod, container.Name)
				if err != nil {
					return fmt.Errorf("couldn't get pod logs: %s", err)
				}
				podLogs = append(podLogs, report.PodLog{
					PodName:       pod.ObjectMeta.Name,
					ContainerName: container.Name,
					PodLogs:       strings.Replace(log, "\"", "", -1),
				})
			}
		}
	}

	debugFile, err := options.fs.OpenFile(reportName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer debugFile.Close()

	err = report.DebugJsonReport(debugFile, report.TemplateData{
		OcpVersion:   options.ocpVersion,
		Subscription: *options.subscription,
		Csv:          options.csv,
		CsvTimeout:   options.csvTimeout,
		CsvEvents:    events,
		PodEvents:    podEvents,
		PodLogs:      podLogs,
	})

	if err != nil {
		return fmt.Errorf("could not generate debug JSON report: %v", err)
	}

	return nil
}

// kubeConfig return kubernetes cluster config
func kubeConfig() (*rest.Config, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		// returned when there is no kubeconfig
		if errors.Is(err, clientcmd.ErrEmptyConfig) {
			return nil, fmt.Errorf("please provide kubeconfig before retrying: %v", err)
		}

		// returned when the kubeconfig has no servers
		if errors.Is(err, clientcmd.ErrEmptyCluster) {
			return nil, fmt.Errorf("malformed kubeconfig. Please check before retrying: %v", err)
		}

		// any other errors getting kubeconfig would be caught here
		return nil, fmt.Errorf("error getting kubeocnfig. Please check before retrying: %v", err)
	}
	return config, nil
}

func k8sClientset() (*kubernetes.Clientset, error) {
	config, err := kubeConfig()
	if err != nil {
		return nil, fmt.Errorf("couldn't get kubeconfig: %s", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("couldn't create new clientset: %s", err)
	}
	return clientset, nil
}

func OperatorPods(ctx context.Context, clientset *kubernetes.Clientset, namespace string) (*corev1.PodList, error) {
	podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("couldn't list pods for operator: %s", err.Error())
	}

	return podList, nil
}

func EventsByNameAndKind(ctx context.Context, clientset *kubernetes.Clientset, name string, kind string, namespace string) (*corev1.EventList, error) {
	events, err := clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{FieldSelector: "involvedObject.name=" + name, TypeMeta: metav1.TypeMeta{Kind: kind}})
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve events: %s", err)
	}
	return events, nil
}

func Logs(ctx context.Context, clientset *kubernetes.Clientset, pod corev1.Pod, container string) (string, error) {
	podLogOpts := corev1.PodLogOptions{Container: container}
	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("error in opening stream: %s", err)
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", fmt.Errorf("error in copy information from podLogs to buf %s", err)
	}
	str := buf.String()

	return str, nil
}
