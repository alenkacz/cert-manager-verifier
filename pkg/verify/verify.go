package verify

import (
	"context"
	"fmt"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type VerifyResult struct {
	Success bool

	DeploymentsSuccess bool
	CertificateSuccess bool

	DeploymentsResult []DeploymentResult
	CertificateError error
}

type CertificateResult struct {
	Success bool
	Error error
}

type Options struct {
	CertManagerNamespace string
}

func Verify(ctx context.Context, config *rest.Config, options *Options) (*VerifyResult, error) {
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to get kubernetes client: %v", err)
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to get kubernetes client: %v", err)
	}

	deployments := DeploymentDefinitionDefault(options.CertManagerNamespace)
	deploymentResult := DeploymentsReady(ctx, kubeClient, deployments)

	result := &VerifyResult{
		Success: false,
		DeploymentsResult: deploymentResult,
	}

	if !allReady(deploymentResult) {
		return result, nil
	}
	result.DeploymentsSuccess = true
	err = WaitForTestCertificate(ctx, dynamicClient)
	if err != nil {
		result.CertificateError = err
	} else {
		result.CertificateSuccess = true
		result.Success = true
	}

	return result, nil
}

func allReady(result []DeploymentResult) bool {
	for _, r := range result {
		if !r.Ready {
			return false
		}
	}
	return true
}
