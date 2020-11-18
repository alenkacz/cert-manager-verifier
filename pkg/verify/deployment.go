package verify

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/polymorphichelpers"
)

func DeploymentReady(kubeClient *kubernetes.Clientset) error {
	// TODO handle non-default namespace
	// TODO make sure these names work also with helm chart installation
	// TODO make sure we support cert-manager that does not have all these deployments
	deploymentReady(kubeClient, "cert-manager", "cert-manager")
	caInjectorDeployment, err :=kubeClient.AppsV1().Deployments("cert-manager").Get(context.TODO(), "cert-manager-cainjector", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error when retrieving cert-manager deployments: %v", err)
	}
	webhookDeployment, err :=kubeClient.AppsV1().Deployments("cert-manager").Get(context.TODO(), "cert-manager-webhook", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error when retrieving cert-manager deployments: %v", err)
	}
}

type UnhealthyError struct {
	
}

func deploymentReady(kubeClient *kubernetes.Clientset, name, namespace string) error {
	statusViewer := &polymorphichelpers.DeploymentStatusViewer{}
	cmDeployment, err := kubeClient.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error when retrieving cert-manager deployments: %v", err)
	}
	unst, err := toUnstructured(cmDeployment)
	if err != nil {
		return fmt.Errorf("error when converting deployment to unstructured: %v", err)
	}
	msg, healthy, err := statusViewer.Status(unst, 0)
	if err != nil {
		return fmt.Errorf("error when evaluating health of deployment: %v", err)
	}
	if !healthy {
		return fmt.Errorf("deployment %v is marked healthy", objUnstructured.GetName()), nil
	}
}

func toUnstructured(obj runtime.Object) (*unstructured.Unstructured, error) {
	unstructMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: unstructMap}, nil
}