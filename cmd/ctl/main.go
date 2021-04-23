package main

import (
	"os"

	"github.com/alenkacz/cert-manager-verifier/pkg/cmd/verify"
)

func main() {
	err := verify.NewCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
