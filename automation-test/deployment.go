package main

import (
	"fmt"
	"log"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
)

//TestDeployment tests deployment creation/deletion
func TestDeployment(client clientset.Interface) {
	log.Println("Creating a deployment...")
	deployment, err := CreateDeployment(client)
	if err != nil {
		// need fix
		log.Fatal(err.Error())
	}

	log.Println(deployment)

	log.Printf("Created the deployment %q.\n", deployment.GetObjectMeta().GetName())
}

// CreateDeployment creates a deployment.
func CreateDeployment(client clientset.Interface) (*appsv1.Deployment, error) {
	zero := int64(0)
	replica := int32(2)
	deploymentSpec := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replica,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "my-deployment",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "my-deployment",
					},
				},
				Spec: apiv1.PodSpec{
					TerminationGracePeriodSeconds: &zero,
					Containers: []apiv1.Container{
						{
							Name:    "myapp-container",
							Image:   "busybox:1.29",
							Command: []string{"/bin/sh"},
							Args:    []string{"-c", "echo The app is running! && sleep 3600"},
						},
					},
					RestartPolicy: apiv1.RestartPolicyAlways,
				},
			},
		},
	}

	deployment, err := client.AppsV1().Deployments(apiv1.NamespaceDefault).Create(deploymentSpec)
	if err != nil {
		return nil, fmt.Errorf("deployment %q Create API error: %v", deploymentSpec.Name, err)
	}
	log.Printf("Waiting deployment %q to complete", deploymentSpec.Name)
	err = WaitForDeploymentComplete(client, deployment)
	if err != nil {
		return nil, fmt.Errorf("deployment %q failed to complete: %v", deploymentSpec.Name, err)
	}
	return deployment, nil
}

// WaitForDeploymentComplete waits for the deployment to complete.
func WaitForDeploymentComplete(c clientset.Interface, d *appsv1.Deployment) error {
	var (
		deployment *appsv1.Deployment
		reason     string
	)

	err := wait.PollImmediate(poll, pollLongTimeout, func() (bool, error) {
		var err error
		deployment, err = c.AppsV1().Deployments(d.Namespace).Get(d.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		// When the deployment status and its underlying resources reach the desired state, we're done
		if DeploymentComplete(d, &deployment.Status) {
			return true, nil
		}

		reason = fmt.Sprintf("deployment status: %#v", deployment.Status)
		log.Println(reason)

		return false, nil
	})

	if err == wait.ErrWaitTimeout {
		err = fmt.Errorf("error waiting timeout: %s", reason)
	}
	if err != nil {
		return fmt.Errorf("error waiting for deployment %q status to match expectation: %v", d.Name, err)
	}
	return nil
}

// DeploymentComplete considers a deployment to be complete once all of its desired replicas
// are updated and available, and no old pods are running.
func DeploymentComplete(deployment *appsv1.Deployment, newStatus *appsv1.DeploymentStatus) bool {
	return newStatus.UpdatedReplicas == *(deployment.Spec.Replicas) &&
		newStatus.Replicas == *(deployment.Spec.Replicas) &&
		newStatus.AvailableReplicas == *(deployment.Spec.Replicas) &&
		newStatus.ObservedGeneration >= deployment.Generation
}
