package verify

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
)

const (
	defaultGroup   = "cert-manager.io"
	defaultVersion = "v1"
)

var namespace = &unstructured.Unstructured{
	Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Namespace",
		"metadata": map[string]interface{}{
			"name": "cert-manager-test",
		},
	},
}

// TODO support also other API versions
// TODO make it possible to execute this inside namespace, not creating one
func WaitForTestCertificate(ctx context.Context, dynamicClient dynamic.Interface, cmVersion string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("Timeout reached: %v", err)
	}
	group, version := getGroupVersion(cmVersion)
	cert := certificate("cert-manager-test", group, version)
	resources := []*unstructured.Unstructured{namespace, issuer("cert-manager-test", group, version), cert}
	defer cleanupTestResources(dynamicClient, resources)

	for _, res := range resources {
		// we need to retry here because cert-manager webhook might not be ready yet
		err := createWithRetry(ctx, res, dynamicClient)
		if err != nil {
			return err
		}
	}
	poller := &certPoller{dynamicClient, cert}
	return wait.PollImmediateUntil(defaultPollInterval, poller.certificateReady, ctx.Done())
}

func getGroupVersion(cmVersion string) (string, string) {
	if strings.HasPrefix(cmVersion, "v1.") {
		return defaultGroup, defaultVersion
	} else {
		return defaultGroup, "v1alpha2"
	}
}

func createWithRetry(ctx context.Context, res *unstructured.Unstructured, dynamicClient dynamic.Interface) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("Timeout reached: %v", ctx.Err())
		default:
			err := createResource(dynamicClient, res)
			if errors.IsAlreadyExists(err) {
				logrus.Debugf("Resource %s already exists \n", res.GetName())
			} else if err != nil {
				logrus.Debugf("Retrying create of resource %s, error: %v\n", res.GetName(), err)
			} else {
				logrus.Debugf("Resource %s created \n", res.GetName())
				return nil
			}
		}
	}
}

type certPoller struct {
	dynamicClient dynamic.Interface
	certificate   *unstructured.Unstructured
}

func (p *certPoller) certificateReady() (bool, error) {
	gvk := p.certificate.GroupVersionKind()
	cert, err := p.dynamicClient.Resource(schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: fmt.Sprintf("%ss", strings.ToLower(gvk.Kind)), // since we know what kinds are we dealing with here, this is OK
	}).Namespace(p.certificate.GetNamespace()).Get(context.TODO(), p.certificate.GetName(), metav1.GetOptions{}, "status")
	if err != nil {
		return false, err
	}
	conditions, exists, err := unstructured.NestedSlice(cert.Object, "status", "conditions")
	if !exists || err != nil {
		return false, err
	}
	for _, c := range conditions {
		reason, found, err := unstructured.NestedString(c.(map[string]interface{}), "type")
		if !found || err != nil {
			return false, err
		}
		if reason == "Ready" {
			status, found, err := unstructured.NestedString(c.(map[string]interface{}), "status")
			if !found || err != nil {
				return false, err
			}
			return status == "True", nil
		}
	}
	return false, nil
}

func createResource(dynamicClient dynamic.Interface, resource *unstructured.Unstructured) error {
	gvk := resource.GroupVersionKind()
	_, err := dynamicClient.Resource(schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: fmt.Sprintf("%ss", strings.ToLower(gvk.Kind)), // since we know what kinds are we dealing with here, this is OK
	}).Namespace(resource.GetNamespace()).Create(context.TODO(), resource, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		logrus.Debugf("resource %s already exists\n", resource.GetName())
	} else if err != nil {
		return fmt.Errorf("error when creating resource %s/%s. %v", resource.GetName(), resource.GetNamespace(), err)
	}
	return nil
}

func deleteResource(dynamicClient dynamic.Interface, resource *unstructured.Unstructured) error {
	gvk := resource.GroupVersionKind()
	err := dynamicClient.Resource(schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: fmt.Sprintf("%ss", strings.ToLower(gvk.Kind)), // since we know what kinds are we dealing with here, this is OK
	}).Namespace(resource.GetNamespace()).Delete(context.TODO(), resource.GetName(), metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		logrus.Debugf("resource %s already deleted\n", resource.GetName())
	} else if err != nil {
		return fmt.Errorf("error when creating resource %s/%s. %v", resource.GetName(), resource.GetNamespace(), err)
	}
	return nil
}

func cleanupTestResources(dynamicClient dynamic.Interface, resources []*unstructured.Unstructured) error {
	for _, res := range resources {
		err := deleteResource(dynamicClient, res)
		if err != nil {
			return err
		}
	}
	return nil
}

func issuer(ns string, group string, apiVersion string) *unstructured.Unstructured {
	apiString := fmt.Sprintf("%s/%s", group, apiVersion)
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiString,
			"kind":       "Issuer",
			"metadata": map[string]interface{}{
				"name":      "test-selfsigned",
				"namespace": ns,
			},
			"spec": map[string]interface{}{
				"selfSigned": map[string]interface{}{},
			},
		},
	}
}

func certificate(ns string, group string, apiVersion string) *unstructured.Unstructured {
	apiString := fmt.Sprintf("%s/%s", group, apiVersion)
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiString,
			"kind":       "Certificate",
			"metadata": map[string]interface{}{
				"name":      "selfsigned-cert",
				"namespace": ns,
			},
			"spec": map[string]interface{}{
				"dnsNames": []string{"example.com"},
				"issuerRef": map[string]interface{}{
					"kind": "Issuer",
					"name": "test-selfsigned",
				},
				"secretName": "selfsigned-cert-tls",
			},
		},
	}
}
