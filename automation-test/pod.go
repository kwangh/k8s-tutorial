package main

import (
	"fmt"
	"log"
	"time"

	apiv1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
)

//TestPod tests pod creation/deletion
func TestPod(client clientset.Interface) {
	log.Println("Creating a pod...")
	pod, err := CreatePod(client)
	// pod is not running
	if err != nil {
		log.Printf("Creating pod error: %v", err.Error())
		if pod != nil {
			log.Println("Deleting the error pod...")
			// delete pod with wait 5 minutes
			err = DeletePodWithWait(client, pod)
			if err != nil {
				log.Fatal(err.Error())
			}

			log.Println("Delete success.")
		}

		return
	}

	log.Printf("Created the pod %q.\n", pod.Name)
	log.Println("Deleting the pod...")
	// delete pod with wait 5 minutes
	err = DeletePodWithWait(client, pod)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Delete success.")
}

// CreatePod creates a busybox image in default namespace
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
					Name:    "myapp-container",
					Image:   "busybox:1.29",
					Command: []string{"/bin/sh"},
					Args:    []string{"-c", "echo The app is running! && sleep 3600"},
				},
			},
		},
	}

	pod, err := client.CoreV1().Pods(apiv1.NamespaceDefault).Create(pod)
	if err != nil {
		return nil, fmt.Errorf("pod Create API error: %v", err)
	}

	// Waiting for pod to be running
	err = WaitForPodNameRunningInNamespace(client, pod.Name, apiv1.NamespaceDefault)
	if err != nil {
		return pod, fmt.Errorf("pod %q is not Running: %v", pod.Name, err)
	}
	// get fresh pod info
	pod, err = client.CoreV1().Pods(apiv1.NamespaceDefault).Get(pod.Name, metav1.GetOptions{})
	if err != nil {
		return pod, fmt.Errorf("pod Get API error: %v", err)
	}
	return pod, nil
}

// WaitForPodNameRunningInNamespace waits default amount of time (PodStartTimeout) for the specified pod to become running.
// Returns an error if timeout occurs first, or pod goes in to failed state.
func WaitForPodNameRunningInNamespace(c clientset.Interface, podName, namespace string) error {
	return WaitTimeoutForPodRunningInNamespace(c, podName, namespace, podStartTimeout)
}

// WaitTimeoutForPodRunningInNamespace waits the given timeout duration for the specified pod to become running.
func WaitTimeoutForPodRunningInNamespace(c clientset.Interface, podName, namespace string, timeout time.Duration) error {
	return wait.PollImmediate(poll, timeout, podRunning(c, podName, namespace))
}

// podRunning checks whether pod is running
func podRunning(client clientset.Interface, podName, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		pod, err := client.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		switch pod.Status.Phase {
		case apiv1.PodRunning:
			return true, nil
		case apiv1.PodFailed, apiv1.PodSucceeded:
			// ErrPodCompleted is returned by PodRunning or PodContainerRunning to indicate that
			// the pod has already reached completed state.
			return false, fmt.Errorf("pod ran to completion")
		}
		return false, nil
	}
}

// DeletePodWithWait deletes the passed-in pod and waits for the pod to be terminated. Resilient to the pod
// not existing.
func DeletePodWithWait(c clientset.Interface, pod *apiv1.Pod) error {
	if pod == nil {
		return nil
	}
	return DeletePodWithWaitByName(c, pod.GetName(), pod.GetNamespace())
}

// DeletePodWithWaitByName deletes the named and namespaced pod and waits for the pod to be terminated. Resilient to the pod
// not existing.
func DeletePodWithWaitByName(c clientset.Interface, podName, podNamespace string) error {
	err := c.CoreV1().Pods(podNamespace).Delete(podName, nil)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return nil // assume pod was already deleted
		}
		return fmt.Errorf("pod Delete API error: %v", err)
	}
	err = WaitForPodNotFoundInNamespace(c, podName, podNamespace, PodDeleteTimeout)
	if err != nil {
		return fmt.Errorf("pod %q was not deleted: %v", podName, err)
	}
	return nil
}

// WaitForPodNotFoundInNamespace returns an error if it takes too long for the pod to fully terminate.
// Unlike `waitForPodTerminatedInNamespace`, the pod's Phase and Reason are ignored. If the pod Get
// api returns IsNotFound then the wait stops and nil is returned. If the Get api returns an error other
// than "not found" then that error is returned and the wait stops.
func WaitForPodNotFoundInNamespace(c clientset.Interface, podName, ns string, timeout time.Duration) error {
	return wait.PollImmediate(poll, timeout, func() (bool, error) {
		_, err := c.CoreV1().Pods(ns).Get(podName, metav1.GetOptions{})
		if apierrs.IsNotFound(err) {
			return true, nil // done
		}
		if err != nil {
			return true, err // stop wait with error
		}
		return false, nil
	})
}
