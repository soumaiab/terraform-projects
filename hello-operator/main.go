package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// A channel acts as our work queue (holds HelloApp names to reconcile)
var workqueue = make(chan string, 100)

func main() {
	kubeconfig := flag.String("kubeconfig", os.Getenv("HOME")+"/.kube/config", "path to kubeconfig file")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	helloAppRes := schema.GroupVersionResource{
		Group:    "apps.example.com",
		Version:  "v1alpha1",
		Resource: "helloapps",
	}

	// --- Worker: processes items from the queue one by one ---
	go func() {
		for name := range workqueue {
			reconcileByName(name, dynClient, clientset, helloAppRes)
		}
	}()

	// --- HelloApp watcher ---
	helloWatcher, err := dynClient.Resource(helloAppRes).Namespace("default").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	go func() {
		for event := range helloWatcher.ResultChan() {
			obj := event.Object.(*unstructured.Unstructured)
			workqueue <- obj.GetName()
		}
	}()

	// --- Secret watcher ---
	secretWatcher, err := clientset.CoreV1().Secrets("default").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	go func() {
		for event := range secretWatcher.ResultChan() {
			secret := event.Object.(*corev1.Secret)
			secretName := secret.GetName()

			// Find HelloApps that reference this secret
			list, err := dynClient.Resource(helloAppRes).Namespace("default").List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Println("Error listing HelloApps:", err)
				continue
			}
			for _, item := range list.Items {
				spec := item.Object["spec"].(map[string]interface{})
				if s, ok := spec["secretName"].(string); ok && s == secretName {
					fmt.Printf("Secret %q changed → enqueue HelloApp %q\n", secretName, item.GetName())
					workqueue <- item.GetName()
				}
			}
		}
	}()

	fmt.Println("Watching HelloApp and Secret resources in 'default' namespace...")

	// --- Stop handling ---
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
	<-stopCh
	fmt.Println("Shutting down...")
}

// Pull the HelloApp again from API before reconciling
func reconcileByName(name string, dynClient dynamic.Interface, clientset *kubernetes.Clientset, helloAppRes schema.GroupVersionResource) {
	obj, err := dynClient.Resource(helloAppRes).Namespace("default").Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error fetching HelloApp %q: %v\n", name, err)
		return
	}
	reconcileHelloApp(obj, clientset)
}

// Actual reconcile logic
func reconcileHelloApp(obj *unstructured.Unstructured, clientset *kubernetes.Clientset) {
	name := obj.GetName()
	spec := obj.Object["spec"].(map[string]interface{})

	// replicas
	replicas := int32(1)
	if r, ok := spec["replicas"].(int64); ok {
		replicas = int32(r)
	}

	// image
	image := "hellogo:latest"
	if i, ok := spec["image"].(string); ok {
		image = i
	}

	// secretName
	secretName := ""
	if s, ok := spec["secretName"].(string); ok {
		secretName = s
	}

	// fetch the Secret
	secret, err := clientset.CoreV1().Secrets("default").Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		fmt.Println("Error fetching Secret:", err)
		return
	}

	// compute checksum
	checksum := hashSecretData(secret.Data)

	fmt.Printf("Reconciling HelloApp %q → replicas=%d, image=%s, secret=%s\n", name, replicas, image, secretName)

	// Desired Deployment
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
					Annotations: map[string]string{
						"secret-checksum": checksum,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "server",
						Image: image,
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
						}},
						Env: []corev1.EnvVar{{
							Name: "APP_MESSAGE",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: secretName,
									},
									Key: "APP_MESSAGE",
								},
							},
						}},
					}},
				},
			},
		},
	}

	// Get latest Deployment before update
	current, err := clientset.AppsV1().Deployments("default").Get(context.TODO(), name, metav1.GetOptions{})
	if err == nil {
		oldChecksum := current.Spec.Template.Annotations["secret-checksum"]
		if oldChecksum != checksum {
			fmt.Printf("Secret %q changed for HelloApp %q → old checksum=%s, new checksum=%s. Triggering rollout.\n",
				secretName, name, oldChecksum, checksum)
		}
		current.Spec = deploy.Spec
		_, err = clientset.AppsV1().Deployments("default").Update(context.TODO(), current, metav1.UpdateOptions{})
		if err != nil {
			fmt.Println("Error updating Deployment:", err)
		} else {
			fmt.Println("Updated Deployment", name)
		}
	} else {
		_, err = clientset.AppsV1().Deployments("default").Create(context.TODO(), deploy, metav1.CreateOptions{})
		if err != nil {
			fmt.Println("Error creating Deployment:", err)
		} else {
			fmt.Println("Created Deployment", name)
		}
	}

	// Ensure Service exists
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": name},
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			}},
		},
	}
	_, err = clientset.CoreV1().Services("default").Get(context.TODO(), name, metav1.GetOptions{})
	if err == nil {
		_, err = clientset.CoreV1().Services("default").Update(context.TODO(), svc, metav1.UpdateOptions{})
		if err != nil {
			fmt.Println("Error updating Service:", err)
		} else {
			fmt.Println("Updated Service", name)
		}
	} else {
		_, err = clientset.CoreV1().Services("default").Create(context.TODO(), svc, metav1.CreateOptions{})
		if err != nil {
			fmt.Println("Error creating Service:", err)
		} else {
			fmt.Println("Created Service", name)
		}
	}
}

// helper: hash secret contents
func hashSecretData(data map[string][]byte) string {
	h := sha256.New()
	for k, v := range data {
		h.Write([]byte(k))
		h.Write(v)
	}
	return hex.EncodeToString(h.Sum(nil))
}
