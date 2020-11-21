package verify

import (
	"fmt"
	"os"

	"github.com/alenkacz/cert-manager-verifier/pkg/verify"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
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

	result := verify.DeploymentsReady(kubeClient, verify.DeploymentDefinitionDefault())
	for _, r := range result {
		fmt.Printf("%s\t%t\t%s\n", r.Name, r.Ready, r.Error)
	}
	return nil
}
