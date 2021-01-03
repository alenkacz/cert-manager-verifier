package verify

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/alenkacz/cert-manager-verifier/pkg/verify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const defaultTimeout = 2 * time.Minute
var defaultInstallationNamespace = "cert-manager"

type Options struct {
	ConfigFlags *genericclioptions.ConfigFlags
	Streams     *genericclioptions.IOStreams
	DebugLogs   bool
	CertManagerNamespace string
}

func NewOptions() *Options {
	opt := &Options{
		ConfigFlags: genericclioptions.NewConfigFlags(true),
		Streams: &genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
	}
	// this is necessary so that the namespace flag is not inherited from ConfigFlags and we can redefine it
	opt.ConfigFlags.Namespace = nil
	return opt
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

	rootCmd.Flags().BoolVar(&options.DebugLogs, "debug", false, "If true, will print out debug logs (default false)")
	rootCmd.Flags().StringVarP(&options.CertManagerNamespace, "namespace", "n", defaultInstallationNamespace, "Namespace in which cert-manager is installed")

	options.ConfigFlags.AddFlags(rootCmd.Flags())
	rootCmd.SetOut(options.Streams.Out)
	rootCmd.SilenceUsage = true
	// TODO add flag to specify CM version and verify version
	// TODO make timeout configurable

	return rootCmd
}

func (o *Options) Execute() error {
	logrus.SetOutput(o.Streams.Out)
	logrus.SetFormatter(SimpleFormatter{})
	if o.DebugLogs {
		logrus.SetLevel(logrus.DebugLevel)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	config, err := o.ConfigFlags.ToRESTConfig()
	if err != nil {
		return fmt.Errorf("unable to get kubernetes rest config: %v", err)
	}

	logrus.Infof("Waiting for deployments in namespace %s:\n", o.CertManagerNamespace)
	result, err := verify.Verify(ctx, config, &verify.Options{
		o.CertManagerNamespace,
	})
	if err != nil {
		return err
	}

	logrus.Infof(formatDeploymentResult(result.DeploymentsResult))

	if !result.DeploymentsSuccess {
		return fmt.Errorf("FAILED! Not all deployments are ready.")
	}
	if result.CertificateError != nil {
		logrus.
			Infof("error when waiting for certificate to be ready: %v", err)
		return err
	}
	logrus.Info("ヽ(•‿•)ノ Cert-manager is READY!")
	return nil
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

type SimpleFormatter struct{}

func (SimpleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(entry.Message), nil
}
