package verify

import (
	"fmt"
	"github.com/alenkacz/cert-manager-verifier/pkg/verify"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"os"
)

type Options struct {
	ConfigFlags *genericclioptions.ConfigFlags
	Streams     *genericclioptions.IOStreams
}

func NewOptions() *Options {
	return &Options{
		ConfigFlags: genericclioptions.NewConfigFlags(true),
		Streams: &genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
	}
}

func NewCmd() *cobra.Command {
	options := NewOptions()

	rootCmd := &cobra.Command{
		Use:   "cert-manager-verifier",
		Short: "Cert Manager verifier helps to verify your cert-manager installation",
		Long: `Cert Manager is used widely in kubernetes clusters and many things depend on it. 
			Unfortunately it's not so easy to know that cert-manager is installed and readiness probes are not
			enough here.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return options.Execute()
		},
	}

	options.ConfigFlags.AddFlags(rootCmd.Flags())
	rootCmd.SetOut(options.Streams.Out)
	rootCmd.SilenceUsage = true
	// TODO add flag to configure cm namespace
	// TODO add flag to specify CM version and verify version

	return rootCmd
}

func (o *Options) Execute() error {
	if o.ConfigFlags.Namespace == nil {
		cmn := "cert-manager"
		o.ConfigFlags.Namespace = &cmn
	}
	config, err := o.ConfigFlags.ToRESTConfig()
	if err != nil {
		return fmt.Errorf("unable to get kubernetes rest config: %v", err)
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("unable to get kubernetes client: %v", err)
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("unable to get kubernetes client: %v", err)
	}

	deployments := verify.DeploymentDefinitionDefault()
	_, _ = fmt.Fprintf(o.Streams.Out, "Waiting for following deployments in namespace %s:\n%s", deployments.Namespace, formatDeploymentNames(deployments.Names))
	result := verify.DeploymentsReady(kubeClient, deployments)
	_, _ = fmt.Fprintf(o.Streams.Out, formatDeploymentResult(result))

	if !allReady(result) {
		return fmt.Errorf("FAILED! Not all deployments are ready.")
	}
	err = verify.WaitForTestCertificate(dynamicClient)
	if err != nil {
		_, _ = fmt.Fprintf(o.Streams.Out, "error when waiting for certificate to be ready: %v", err)
		return err
	}
	_, _ = fmt.Fprint(o.Streams.Out, "ヽ(•‿•)ノ Cert-manager is READY!")
	return nil
}

func allReady(result []verify.DeploymentResult) bool {
	for _, r := range result {
		if !r.Ready {
			return false
		}
	}
	return true
}

func formatDeploymentNames(names []string) string {
	var formattedNames string
	for _, n := range names {
		formattedNames += fmt.Sprintf("\t- %s\n", n)
	}
	return formattedNames

}

func formatDeploymentResult(result []verify.DeploymentResult) string {
	var formattedResult string
	for _, r := range result {
		if r.Ready {
			formattedResult += fmt.Sprintf("Deployment %s READY! ヽ(•‿•)ノ\n", r.Name)
		} else {
			formattedResult += fmt.Sprintf("Deployment %s not ready. Reason: %s\n", r.Name, r.Error.Error())
		}
	}
	return formattedResult
}
