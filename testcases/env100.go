package testcases

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/iftechio/vk-test/suite"
)

func init() {
	t := &env100{}
	register(t)
}

type env100 struct {
}

func (c *env100) Name() string {
	return "container can have more than 100 env vars"
}

func (c *env100) Test(ctx context.Context, s *suite.Suite) error {
	const (
		podName = "100env"
		n       = 110
	)
	envs := make([]corev1.EnvVar, 0, n)
	for i := 0; i < n; i++ {
		envs = append(envs, corev1.EnvVar{
			Name:  fmt.Sprintf("key%d", i),
			Value: fmt.Sprintf("value%d", i),
		})
	}
	ct := corev1.Container{
		Name:  podName,
		Image: "alpine:3.10",
		Env:   envs,
	}
	_, err := s.CoreV1().Pods(metav1.NamespaceDefault).Create(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: corev1.PodSpec{
			Containers:    []corev1.Container{ct},
			RestartPolicy: corev1.RestartPolicyNever,
			NodeName:      s.NodeName(),
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
	return nil
}
