package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	apiv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Action int

const (
	SleepAction Action = iota
	WakeupAction
)

func mustAuthenticate() *kubernetes.Clientset {
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
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// KubernetesClient defines a struct capable of interacting with a Kubernetes cluster
type KubernetesClient struct {
	client *kubernetes.Clientset
}

// IngressRetriever defines the capability of retrieving a map of ingresses
type IngressRetriever interface {
	RetrieveIngresses(ctx context.Context) map[string]bool
}

// DeploymentChecker defines the behavior of checking if a given deployment exists
type DeploymentChecker interface {
	CheckDeployment(ctx context.Context, name string, namespace string) bool
}

type checkPatchReceiver interface {
	DeploymentChecker
	IngressRetriever
	PatchDeployer
}

// PatchDeployer describes the ability to patch a Kubernetes deployer
type PatchDeployer interface {
	PatchDeployment(ctx context.Context, app string, action Action) error
}

func (k KubernetesClient) RetrieveIngresses(ctx context.Context) map[string]bool {
	logrus.Info("Retrieving ingress list")
	ingresses := make(map[string]bool, 0)
	ingressList, err := k.client.ExtensionsV1beta1().Ingresses(metav1.NamespaceAll).List(apiv1.ListOptions{
		LabelSelector: "downscaler.active=true",
	})
	if err != nil {
		logrus.Errorf("Could not retrieve cluster ingresses. Err: %v", err)
		return ingresses
	}
	for _, ingress := range ingressList.Items {
		name := fmt.Sprintf("%s/%s", ingress.Namespace, ingress.Name)
		ingresses[name] = true
	}
	logrus.Infof("AllIngresses total:%d", len(ingresses))
	return ingresses
	// return map[string]bool{
	// 	"ac-identity/acidentity-staging": true, "academy/academy-production": true,
	// 	"acdc/acdcholiday-staging": true, "acdc/acdcrequest-staging": true,
	// 	"acdc/acdctimesheet-staging": true, "acdc/acdctravel-staging": true,
	// 	"acdc/acdcvacation-staging": true, "acdc-legacy/acdclegacy-staging": true,
	// }
}

func (k KubernetesClient) CheckDeployment(ctx context.Context, name string, namespace string) bool {
	apps := k.client.AppsV1()
	_, err := apps.Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		switch t := err.(type) {
		default:
			logrus.Errorf("Some err %v", err)
			return false
		case *kuberr.StatusError:
			if t.ErrStatus.Reason == "NotFound" {
				logrus.Infof("Skiping %s/%s. Could not find a deployment with that name", namespace, name)
				return false
			}
		}
	}
	return true
}

func (k KubernetesClient) PatchDeployment(ctx context.Context, app string, action Action) error {
	var desiredReplicas int32

	i := Ingress(app)
	logrus.Infof("Trying to scale deployment %v", i)
	deployment, err := k.client.AppsV1().Deployments(i.GetNamespace()).Get(i.GetName(), metav1.GetOptions{})
	if err != nil {
		logrus.Errorf("Could not retrieve deployment %v", err)
		return err
	}
	replicas := *deployment.Spec.Replicas
	switch action {
	case WakeupAction:
		if replicas >= 1 {
			logrus.Infof("Skipping scaling app %v since the app already has %d replicas", i, replicas)
			return nil
		}
		desiredReplicas = 1
		logrus.Infof("Scaling app %v back to 1 replica", i)
	case SleepAction:
		if replicas == 0 {
			logrus.Infof("Skipping scale down app %v since it already has 0 replicas", i)
			return nil
		}
		desiredReplicas = 0
	}
	deployment.Spec.Replicas = int32Ptr(desiredReplicas)
	_, err = k.client.AppsV1().Deployments(i.GetNamespace()).Update(deployment)
	return err
}

func int32Ptr(i int32) *int32 { return &i }
