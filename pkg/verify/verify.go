package verify

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type VerifyResult struct {
	Success bool

	DeploymentsSuccess bool
	CertificateSuccess bool

	DeploymentsResult []DeploymentResult
	CertificateError  error
}

type CertificateResult struct {
	Success bool
	Error   error
}

type Options struct {
	CertManagerNamespace string
	DeploymentPrefix     string
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

	deployments := DeploymentDefinitionDefault(options.CertManagerNamespace, options.DeploymentPrefix)
	deploymentResult := DeploymentsReady(ctx, kubeClient, deployments)

	result := &VerifyResult{
		Success:           false,
		DeploymentsResult: deploymentResult,
	}

	if !allReady(deploymentResult) {
		return result, nil
	}
	result.DeploymentsSuccess = true

	cmVersion := version(deploymentResult)

	logrus.Debugf("cert-manager version: %s \n", cmVersion)
	err = WaitForTestCertificate(ctx, dynamicClient, cmVersion)
	if err != nil {
		result.CertificateError = err
	} else {
		result.CertificateSuccess = true
		result.Success = true
	}

	return result, nil
}

func version(result []DeploymentResult) string {
	for _, r := range result {
		if r.Version != "" {
			return r.Version
		}
	}
	return ""
}

func allReady(result []DeploymentResult) bool {
	for _, r := range result {
		if r.Status == NotReady || (r.Status == NotFound && r.Deployment.Required) {
			return false
		}
	}
	return true
}
