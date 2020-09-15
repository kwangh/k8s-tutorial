package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	podsClient := clientset.CoreV1().Pods(apiv1.NamespaceDefault)
	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "myapp-pod",
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

	fmt.Println("Creating pod...")
	result, err := podsClient.Create(pod)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
