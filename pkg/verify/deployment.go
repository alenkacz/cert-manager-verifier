package verify

import (
	"context"
	"fmt"
	"time"

	dp "github.com/novln/docker-parser"
	"k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/polymorphichelpers"
)

type DeploymentDefinition struct {
	Namespace   string
	Deployments []Deployment
}

type Deployment struct {
	Name     string
	Required bool
}

func DeploymentDefinitionDefault(namespace string) DeploymentDefinition {
	// TODO make sure these Deployments work also with helm chart installation
	// TODO make sure we support cert-manager that does not have all these deployments
	return DeploymentDefinition{
		Namespace:   namespace,
		Deployments: []Deployment{{"cert-manager", true}, {"cert-manager-cainjector", false}, {"cert-manager-webhook", false}},
	}
}

type DeploymentResult struct {
	Deployment Deployment
	Status     Status
	Error      error
	Version    string
}

type Status int

const (
	NotReady Status = iota
	Ready
	NotFound
)

// TODO make this configurable
// TODO have a global timeout for all deployments
const defaultPollInterval = 100 * time.Millisecond

func DeploymentsReady(ctx context.Context, kubeClient *kubernetes.Clientset, deployments DeploymentDefinition) []DeploymentResult {
	ctx.Deadline()
	result := []DeploymentResult{}
	for _, d := range deployments.Deployments {
		if err := ctx.Err(); err != nil {
			dr := DeploymentResult{
				Deployment: d,
				Error:      fmt.Errorf("Timeout reached: %v", err),
			}
			result = append(result, dr)
			continue
		}
		dr := DeploymentResult{
			Deployment: d,
			Status:     Ready,
		}
		deployment, err := kubeClient.AppsV1().Deployments(deployments.Namespace).Get(context.TODO(), d.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			dr.Status = NotFound
			result = append(result, dr)
			continue
		}
		if d.Required {
			c := deployment.Spec.Template.Spec.Containers
			if len(c) > 0 {
				r, err := dp.Parse(c[0].Image)
				if err == nil {
					dr.Version = r.Tag()
				}
			}
		}
		poller := &poller{kubeClient, d, deployments.Namespace}
		err = wait.PollImmediateUntil(defaultPollInterval, poller.deploymentReady, ctx.Done())
		if err != nil {
			dr.Status = NotReady
			dr.Error = err
		}
		result = append(result, dr)
	}
	return result
}

type poller struct {
	kubeClient *kubernetes.Clientset
	deployment Deployment
	namespace  string
}

func (p *poller) deploymentReady() (bool, error) {
	statusViewer := &polymorphichelpers.DeploymentStatusViewer{}
	cmDeployment, err := p.kubeClient.AppsV1().Deployments(p.namespace).Get(context.TODO(), p.deployment.Name, metav1.GetOptions{})
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
