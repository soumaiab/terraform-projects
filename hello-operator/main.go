package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
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
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
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
	watcher, err := dynClient.Resource(helloAppRes).Namespace("default").Watch(context.TODO(), metav1.ListOptions{})
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
			name := obj.GetName()
			spec := obj.Object["spec"].(map[string]interface{})

			// pull replicas (default 1 if nil)
			replicas := int32(1)
			if r, ok := spec["replicas"].(int64); ok {
				replicas = int32(r)
			}

			// pull image
			image := "hellogo:latest"
			if i, ok := spec["image"].(string); ok {
				image = i
			}

			fmt.Printf("Reconciling HelloApp %s → replicas=%d, image=%s\n", name, replicas, image)

			// desired Deployment
			deploy := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: "default",
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &replicas,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": name},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": name},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:  "server",
								Image: image,
								Ports: []corev1.ContainerPort{{
									ContainerPort: 8080,
								}},
							}},
						},
					},
				},
			}

			// Check if Deployment exists
			existing, err := clientset.AppsV1().Deployments("default").Get(context.TODO(), name, metav1.GetOptions{})
			if err == nil {
				// update
				existing.Spec = deploy.Spec
				_, err = clientset.AppsV1().Deployments("default").Update(context.TODO(), existing, metav1.UpdateOptions{})
				if err != nil {
					fmt.Println("Error updating Deployment:", err)
				} else {
					fmt.Println("Updated Deployment", name)
				}
			} else {
				// create
				_, err = clientset.AppsV1().Deployments("default").Create(context.TODO(), deploy, metav1.CreateOptions{})
				if err != nil {
					fmt.Println("Error creating Deployment:", err)
				} else {
					fmt.Println("Created Deployment", name)
				}
			}

		}
	}()

	<-stopCh
	fmt.Println("Shutting down...")
}
