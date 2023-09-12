// Package main provides the main entry point for the kubefixtures command.
package main

import (
	"os"

	"github.com/RealHarshThakur/kubefixtures/cmd/load"
	"github.com/RealHarshThakur/kubefixtures/cmd/transition"

	"github.com/spf13/cobra"
)

var (
	kubeconfig string
)

var rootCmd = &cobra.Command{
	Use:  "kubefixtures",
	Long: `kubefixtures is a tool for loading fixtures into a Kubernetes cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func main() {
	Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "k", "~/.kube/config", "Path to kubeconfig file (default: ~/.kube/config)")
	rootCmd.AddCommand(load.LoadCmd)
	rootCmd.AddCommand(transition.TransitionCmd)
}
