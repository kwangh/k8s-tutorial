package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

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
		log.Fatal(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Creating a pod...")
	pod, err := CreatePod(clientset)
	// pod is not running
	if err != nil {
		log.Printf("Creating pod error: %v", err.Error())
		if pod != nil {
			log.Println("Deleting the error pod...")
			// delete pod with wait 5 minutes
			err = DeletePodWithWait(clientset, pod)
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
	err = DeletePodWithWait(clientset, pod)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Delete success.")
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
