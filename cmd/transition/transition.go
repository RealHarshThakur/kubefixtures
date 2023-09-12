package transition

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"

	"github.com/RealHarshThakur/kubefixtures/pkg"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
)

var (
	path           string
	transitionTime int
)

// TransitionCmd is the command for transitioning fixtures
var TransitionCmd = &cobra.Command{
	Use:     "transition",
	Aliases: []string{"change"},
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

		transition(ctx, dynamic, os.Stdout)
	},
}

func transition(ctx context.Context, dynamic dynamic.Interface, o io.Writer) {
	fr := pkg.SetupFixtureLoader(dynamic, pkg.SetupLogging(o))
	log := pkg.SetupLogging(os.Stdout)
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

	err = fr.StatusLoad(ctx, ri, obj)
	if err != nil {
		os.Exit(1)
	}

	log.Printf("Successfully transitioned fixture(s)")
}

func init() {
	TransitionCmd.Flags().StringVarP(&path, "file-path", "f", "", "file path to the fixture to load")
}
