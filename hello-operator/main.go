package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	// "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Load kubeconfig (from ~/.kube/config by default)
	kubeconfig := flag.String("kubeconfig", os.Getenv("HOME")+"/.kube/config", "path to kubeconfig file")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	// Dynamic client (works with CRDs)
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// Tell client which resource we want
	helloAppRes := schema.GroupVersionResource{
		Group:    "apps.example.com",
		Version:  "v1alpha1",
		Resource: "helloapps", // plural name from the CRD
	}

	// Watch for HelloApp objects
	watcher, err := client.Resource(helloAppRes).Namespace("default").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Println("Watching HelloApp resources in 'default' namespace...")

	// Handle signals so we can stop cleanly
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for event := range watcher.ResultChan() {
			obj := event.Object.(*unstructured.Unstructured)
			fmt.Printf("Event: %s, Name: %s, Spec: %+v\n",
				event.Type,
				obj.GetName(),
				obj.Object["spec"],
			)
		}
	}()

	<-stopCh
	fmt.Println("Shutting down...")
}
