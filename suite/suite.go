package suite

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type Suite struct {
	kubernetes.Interface
	nodeName   string
	restConfig *rest.Config
}

func (s *Suite) WaitUntilPodSteady(ctx context.Context, namespace, name string) error {
	watcher, err := s.Interface.CoreV1().Pods(namespace).Watch(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
	})
	if err != nil {
		return err
	}
	defer watcher.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ev, ok := <-watcher.ResultChan():
			if !ok {
				return fmt.Errorf("channel closed")
			}
			pod := ev.Object.(*corev1.Pod)
			switch pod.Status.Phase {
			case corev1.PodRunning, corev1.PodSucceeded, corev1.PodFailed:
				return nil
			}
		}
	}
}

func (s *Suite) DeletePod(namespace, name string) error {
	policy := metav1.DeletePropagationForeground
	return s.Interface.CoreV1().Pods(namespace).Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &policy,
	})
}

func (s *Suite) ExecInContainer(namespace, pod, container string, command []string) (stdout, stderr string, err error) {
	req := s.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("exec").
		Param("container", container)
	req.VersionedParams(&corev1.PodExecOptions{
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		Container: pod,
		Command:   command,
	}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(s.restConfig, "POST", req.URL())
	if err != nil {
		return "", "", err
	}
	var (
		outBuf, errBuf bytes.Buffer
	)
	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: &outBuf,
		Stderr: &errBuf,
	})
	if err != nil {
		return "", "", err
	}
	return strings.TrimSpace(outBuf.String()), strings.TrimSpace(errBuf.String()), nil
}

func (s *Suite) NodeName() string {
	return s.nodeName
}

func New(restCfg *rest.Config, nodeName string) (*Suite, error) {
	cliset, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, err
	}
	s := &Suite{
		Interface:  cliset,
		restConfig: restCfg,
		nodeName:   nodeName,
	}
	return s, nil
}
