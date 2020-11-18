package main

import "github.com/alenkacz/cert-manager-verifier/pkg/cmd/verify"

func main() {
	verify.NewCmd().Execute()
}
