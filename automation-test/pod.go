package main

import (
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

const (
	// podStartTimeout is how long to wait for the pod to be started.
	// Initial pod start can be delayed O(minutes) by slow docker pulls.
	// TODO: Make this 30 seconds once #4566 is resolved.
	podStartTimeout = 5 * time.Minute

	// poll is how often to poll pods, nodes and claims.
	poll = 2 * time.Second
)

// CreatePod with busybox image in default namespace
func CreatePod(client clientset.Interface) (*apiv1.Pod, error) {
	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "myapp",
			Labels: map[string]string{
				"app": "myapp",
			},
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Name:  "myapp-container",
					Image: "busybox:1.28",
					Command: []string{
						"sh",
						"-c",
						"echo The app is running! && sleep 3600",
					},
				},
			},
		},
	}

	pod, err := client.CoreV1().Pods(apiv1.NamespaceDefault).Create(pod)
	if err != nil {
		return nil, fmt.Errorf("pod Create API error: %v", err)
	}

	// get fresh pod info
	pod, err = client.CoreV1().Pods(apiv1.NamespaceDefault).Get(pod.Name, metav1.GetOptions{})
	if err != nil {
		return pod, fmt.Errorf("pod Get API error: %v", err)
	}
	return pod, nil
}
