# cert-manager-verifier

Helps you properly wait for [cert-manager](https://github.com/jetstack/cert-manager) installation to be ready to use.

All versions of cert-manager are supported down to 0.12.

## Motivation

Pretty much every kubernetes installation nowadays depends on cert-manager for certificate provisioning. But at the same time, I've seen a lot of flakiness in several projects that were caused by inproperly implemented wait on cert-manager to be ready.

Currently the right way to do this is documented in [installation guide](https://cert-manager.io/docs/installation/kubernetes/#verifying-the-installation). It consists of several steps and typically in a CI or automated environment, you don't want to execute those manually and that's where this project aims to help.

## Usage

There are two ways how to use the verifier - CLI and as a go library and your choice depends on what your cluster installation tooling looks like.

### CLI integration

The CLI is just another binary you call from your CI/CD pipeline or cluster installation scripts. When it returns exitcode 0 you know cert-manager is ready to use.

It expects kubeconfig/user that is allowed to create namespace in the default settings (it creates and cleans up cert-manager-test ns to deploy the test certificates). See `cm-verifier --help` for more information about how to configure this.

```shell script
    ./cm-verifier
    Waiting for following deployments in namespace cert-manager:
    	- cert-manager
    	- cert-manager-cainjector
    	- cert-manager-webhook
    Deployment cert-manager READY! ヽ(•‿•)ノ
    Deployment cert-manager-cainjector READY! ヽ(•‿•)ノ
    Deployment cert-manager-webhook READY! ヽ(•‿•)ノ
    Resource cert-manager-test created
    Resource test-selfsigned created
    Resource selfsigned-cert created
    ヽ(•‿•)ノ Cert-manager is READY!%
```

You can configure what the CLI does via flags:
```
--debug 'print out debug logs as well'
--namespace 'namespace into which cert-manager is installed'
--timeout 'set timeout after which we give up waiting for cert-manager'
```

### As a library

For a full example see [examples/main.go](examples/main.go)

```
import "github.com/alenkacz/cert-manager-verifier/pkg/verify"

...

result, err := verify.Verify(ctx, config, &verify.Options{CertManagerNamespace: "cert-manager"})
if err != nil {
    log.Fatal(err)
}
if result.Success {
    fmt.Println("Success!!!")
} else {
    fmt.Println("Failure :-(")
}
```
