package main

import (
	"context"
	"fmt"
	"github.com/alenkacz/cert-manager-verifier/pkg/verify"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"time"
	"log"
)

const defaultTimeout = 2 * time.Minute

func main() {
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
	kubeconfig := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	result, err := verify.Verify(ctx, config, &verify.Options{CertManagerNamespace: "cert-manager"})
	if err != nil {
		log.Fatal(err)
	}
	if result.Success {
		fmt.Println("Success!!!")
	} else {
		fmt.Println("Failure :-(")
	}
}