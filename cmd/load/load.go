package load

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"

	"kubefixtures/pkg"
)

var (
	path           string
	status         string
	transitionTime int
)

// LoadCmd is the command for loading fixtures
var LoadCmd = &cobra.Command{
	Use:     "load",
	Aliases: []string{"create"},
	Short:   "Load a fixture into the cluster",
	Long:    `Load a fixture into the cluster `,
	Example: "kubefixtures load -f example.yaml",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		log := pkg.SetupLogging(os.Stdout)
		kubeconfig, err := cmd.Root().PersistentFlags().GetString("kubeconfig")
		if err != nil {
			log.Fatalf("Failed to get kubeconfig: %v", err)
		}

		dynamic, err := pkg.SetupDynamicClient(&kubeconfig)
		if err != nil {
			log.Fatalf("Failed to setup dynamic client: %v", err)
		}

		load(ctx, dynamic, os.Stdout)

	},
}

func load(ctx context.Context, dynamic dynamic.Interface, o io.Writer) {
	fr := pkg.SetupFixtureLoader(dynamic, pkg.SetupLogging(o))

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read YAML file: %v", err)
	}

	// Convert YAML to unstructured object
	// TODO: Support loading multiple objects from a single file and multiple files
	var obj unstructured.Unstructured
	if err := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 100).Decode(&obj); err != nil {
		log.Fatalf("Failed to decode YAML to unstructured object: %v", err)
	}

	ri := pkg.GetResourceInfo(obj)
	err = fr.CreateResourceDynamically(ctx, ri, obj)
	if err != nil {
		log.Fatal("Failed. To ensure what you have in fixtures is same as what you have in clusters, do a `kubectl delete -f <file>`, then try again")
	}

	// Load the status from YAML into the cluster
	err = fr.StatusLoad(ctx, ri, obj)
	if err != nil {
		os.Exit(1)
	}

	log.Printf("Successfully loaded fixture(s)")
}

func convertStatusToKeyValue(status string) (string, interface{}) {
	s := strings.Split(status, "=")
	if len(s) != 2 {
		log.Fatalf("Invalid status format. Expected 'key=value'")
	}
	return s[0], s[1]
}

func init() {
	LoadCmd.Flags().StringVarP(&path, "file-path", "f", "", "file path to the fixture to load")
	LoadCmd.Flags().IntVarP(&transitionTime, "time", "t", 10, "time to wait(in seconds) for the status field to change")
}
