package verify

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
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
	Error   error
}

// TODO make this configurable
// TODO have a global timeout for all deployments
const defaultPollInterval = 100 * time.Millisecond

func DeploymentsReady(ctx context.Context, kubeClient *kubernetes.Clientset, deployments DeploymentDefinition) []DeploymentResult {
	ctx.Deadline()
	result := []DeploymentResult{}
	for _, d := range deployments.Names {
		if err := ctx.Err(); err != nil {
			dr := DeploymentResult{
				Name:  d,
				Error: fmt.Errorf("Timeout reached: %v", err),
			}
			result = append(result, dr)
			continue
		}

		poller := &poller{kubeClient, d, deployments.Namespace}
		err := wait.PollImmediateUntil(defaultPollInterval, poller.deploymentReady, ctx.Done())
		dr := DeploymentResult{
			Name:  d,
			Ready: true,
		}
		if err != nil {
			dr.Ready = false
			dr.Error = err
		}
		result = append(result, dr)
	}
	return result
}

type poller struct {
	kubeClient *kubernetes.Clientset
	name       string
	namespace  string
}

func (p *poller) deploymentReady() (bool, error) {
	statusViewer := &polymorphichelpers.DeploymentStatusViewer{}
	cmDeployment, err := p.kubeClient.AppsV1().Deployments(p.namespace).Get(context.TODO(), p.name, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("error when retrieving cert-manager deployments: %v", err)
	}
	unst, err := toUnstructured(cmDeployment)
	if err != nil {
		return false, fmt.Errorf("error when converting deployment to unstructured: %v", err)
	}
	_, ready, err := statusViewer.Status(unst, 0)
	if err != nil {
		return false, nil
	}
	return ready, nil
}

func toUnstructured(obj runtime.Object) (*unstructured.Unstructured, error) {
	unstructMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: unstructMap}, nil
}
