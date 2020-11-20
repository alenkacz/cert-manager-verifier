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

type DeploymentDefinition struct {
	Namespace string
	Names     []string
}

func DeploymentDefinitionDefault() DeploymentDefinition {
	// TODO make sure these Names work also with helm chart installation
	// TODO make sure we support cert-manager that does not have all these deployments
	return DeploymentDefinition{
		Namespace: "cert-manager",
		Names:     []string{"cert-manager", "cert-manager-cainjector", "cert-manager-webhook"},
	}
}

type DeploymentResult struct {
	Name    string
	Ready   bool
	Message string
	Error   error
}

func DeploymentReady(kubeClient *kubernetes.Clientset, deployments DeploymentDefinition) []DeploymentResult {
	result := []DeploymentResult{}
	for _, d := range deployments.Names {
		// TODO add wait time and timeout
		result = append(result, deploymentReady(kubeClient, d, deployments.Namespace))
	}
	return result
}

func deploymentReady(kubeClient *kubernetes.Clientset, name, namespace string) DeploymentResult {
	statusViewer := &polymorphichelpers.DeploymentStatusViewer{}
	result := DeploymentResult{
		Name:  name,
		Ready: false,
	}
	cmDeployment, err := kubeClient.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		result.Error = fmt.Errorf("error when retrieving cert-manager deployments: %v", err)
		return result
	}
	unst, err := toUnstructured(cmDeployment)
	if err != nil {
		result.Error = fmt.Errorf("error when converting deployment to unstructured: %v", err)
		return result
	}
	result.Message, result.Ready, result.Error = statusViewer.Status(unst, 0)
	return result
}

func toUnstructured(obj runtime.Object) (*unstructured.Unstructured, error) {
	unstructMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: unstructMap}, nil
}
