package testcases

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/iftechio/vk-test/suite"
)

func init() {
	t := &persistEmptyDir{}
	register(t)
}

type persistEmptyDir struct {
}

func (c *persistEmptyDir) Name() string {
	return "emptydir"
}

func (c *persistEmptyDir) Description() string {
	return "emptyDir should not be cleared after updating pod"
}

func (c *persistEmptyDir) Test(ctx context.Context, s *suite.Suite) error {
	const podName = "persistent-empty-dir"
	_, err := s.CoreV1().Pods(metav1.NamespaceDefault).Create(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:    podName,
				Image:   "alpine:3.10",
				Command: []string{"sh", "-c", "sleep 86400"},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "data",
						MountPath: "/data",
					},
				},
			}},
			RestartPolicy: corev1.RestartPolicyNever,
			NodeName:      s.NodeName(),
			Volumes: []corev1.Volume{
				{
					Name: "data",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("create pod: %s", err)
	}
	defer s.DeletePod(metav1.NamespaceDefault, podName)
	err = s.WaitUntilPodSteady(ctx, metav1.NamespaceDefault, podName)
	if err != nil {
		return fmt.Errorf("wait for pod creation: %s", err)
	}
	_, _, err = s.ExecInContainer(metav1.NamespaceDefault, podName, podName, []string{"sh", "-c", "date > /data/1.txt"})
	if err != nil {
		return fmt.Errorf("exec: %s", err)
	}
	pod, err := s.CoreV1().Pods(metav1.NamespaceDefault).Get(podName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	pod.Spec.Containers[0].Image = "alpine:3.8"
	_, err = s.CoreV1().Pods(metav1.NamespaceDefault).Update(pod)
	if err != nil {
		return fmt.Errorf("update pod: %s", err)
	}
	time.Sleep(time.Second * 5)
	err = s.WaitUntilPodSteady(ctx, metav1.NamespaceDefault, podName)
	if err != nil {
		return fmt.Errorf("wait for pod: %s", err)
	}
	_, stderr, err := s.ExecInContainer(metav1.NamespaceDefault, podName, podName, []string{"sh", "-c", "cat /data/1.txt"})
	if err != nil {
		return fmt.Errorf("exec: %s: %s", err, stderr)
	}
	return nil
}
