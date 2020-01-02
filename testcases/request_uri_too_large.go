package testcases

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/iftechio/vk-test/suite"
)

func init() {
	t := &requestURITooLarge{}
	register(t)
}

type requestURITooLarge struct {
}

func (c *requestURITooLarge) Name() string {
	return "request uri too large"
}

func (c *requestURITooLarge) Description() string {
	return "tengine should not return 414 request uri too large"
}

func (c *requestURITooLarge) Test(ctx context.Context, s *suite.Suite) error {
	const (
		podName      = "request-uri-too-large"
		envLen       = 100
		containerLen = 11
	)
	// 模拟真实环境env长度
	envs := make([]corev1.EnvVar, 0, envLen)
	for i := 0; i < 100; i++ {
		envs = append(envs, corev1.EnvVar{
			Name:  fmt.Sprintf("MockKey%d", i),
			Value: fmt.Sprintf("ThisIsALongMockString%d", i),
		})
	}
	// 模拟真实环境container数量
	var containers []corev1.Container
	for i := 0; i < containerLen; i++ {
		containers = append(containers, corev1.Container{
			Name:  fmt.Sprintf("test%d", i),
			Image: "busybox",
			Env:   envs,
		})
	}
	_, err := s.CoreV1().Pods(metav1.NamespaceDefault).Create(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: metav1.NamespaceDefault,
			Annotations: map[string]string{
				"k8s.aliyun.com/eci-cpu":    "1",
				"k8s.aliyun.com/eci-memory": "2Gi",
			},
		},
		Spec: corev1.PodSpec{
			Containers:    containers,
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
