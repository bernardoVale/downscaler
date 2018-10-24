package kube

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bernardoVale/downscaler/internal/types"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type (
	// KubernetesClient defines a struct capable of interacting with a Kubernetes cluster
	KubernetesClient struct {
		client *kubernetes.Clientset
	}

	// Lister is an interface that wraps a base List method
	Lister interface {
		List() map[string]bool
	}

	// PatchLister defines the capability of receiving a list of ingresses
	// and patch a deployment spec
	PatchLister interface {
		Lister
		Scaler
	}

	// Getter is an interface that wraps a base Get method
	GetDeployment interface {
		Get(name string, namespace string) (deployment *appsv1.Deployment, err error)
	}

	// Scaler describes the ability of scaling an application
	Scaler interface {
		Scale(namespace string, name string, scaleType types.ScaleType) error
	}

	GetScaler interface {
		GetDeployment
		Scaler
	}
)

func localAuth(kubeconfig *string) (*kubernetes.Clientset, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func inClusterAuth() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func auth() (*kubernetes.Clientset, error) {
	home := homeDir()
	kubeConfigPath := filepath.Join(home, ".kube", "config")
	if _, err := os.Stat(kubeConfigPath); os.IsNotExist(err) {
		// if err != nil {
		// 	return nil, err
		// }
		return inClusterAuth()
	}
	return localAuth(&kubeConfigPath)
}

// NewKubernetesClient
func NewKubernetesClient() (*KubernetesClient, error) {
	clientset, err := auth()
	if err != nil {
		return nil, err
	}
	return &KubernetesClient{client: clientset}, nil
}

func (c *KubernetesClient) Watch() {
	events, err := c.client.CoreV1().ConfigMaps("default").Watch(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for {
		select {
		case ev, ok := <-events.ResultChan():
			if !ok {
				break
			}
			logrus.Infof("Object:%s, Type: %s", ev.Object, ev.Type)
		}
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// List returns a List of ingresses from all namespaces that has
// downscaler enabled
func (c KubernetesClient) List() map[string]bool {
	l := logrus.WithField("method", "kube:KubernetesClient:List")

	l.Info("Retrieving ingress list")
	ingresses := make(map[string]bool, 0)
	ingressList, err := c.client.ExtensionsV1beta1().Ingresses(metav1.NamespaceAll).List(apiv1.ListOptions{
		LabelSelector: "downscaler.active=true",
	})
	if err != nil {
		l.Errorf("Could not retrieve cluster ingresses. Err: %v", err)
		return ingresses
	}
	for _, ingress := range ingressList.Items {
		name := fmt.Sprintf("%s/%s", ingress.Namespace, ingress.Name)
		ingresses[name] = true
	}
	l.Infof("AllIngresses total:%d", len(ingresses))
	return ingresses
}

// Get KubernetesClient  retrieves a deployment spec
func (c KubernetesClient) Get(name string, namespace string) (deployment *appsv1.Deployment, err error) {
	apps := c.client.AppsV1()
	return apps.Deployments(namespace).Get(name, metav1.GetOptions{})
}

// Scale up or down a Kubernetes deployment
func (c KubernetesClient) Scale(namespace string, name string, scaleType types.ScaleType) error {
	l := logrus.WithFields(logrus.Fields{
		"method": "kube:KubernetesClient:Patch",
		"app":    fmt.Sprintf("%s/%s", namespace, name),
	})
	var desiredReplicas int32

	l.Info("Trying to scale deployment")
	deployment, err := c.client.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		l.WithError(err).Errorf("Could not retrieve deployment")
		return err
	}
	replicas := *deployment.Spec.Replicas
	switch scaleType {
	case types.ScaleUp:
		if replicas >= 1 {
			l.Infof("Skipping scaling app since the app already has %d replicas", replicas)
			return nil
		}
		desiredReplicas = 1
		l.Info("Scaling app back to 1 replica")
	case types.ScaleDown:
		if replicas == 0 {
			l.Info("Skipping scale down app since it already has 0 replicas")
			return nil
		}
		desiredReplicas = 0
	}
	l.Infof("Set desired replicas to %d", desiredReplicas)
	deployment.Spec.Replicas = int32Ptr(desiredReplicas)

	deployment, err = c.client.AppsV1().Deployments(namespace).Update(deployment)
	if err != nil {
		l.WithError(err).Errorf("Could not patch deployment.")
		return err
	}
	return nil
}

func int32Ptr(i int32) *int32 { return &i }
