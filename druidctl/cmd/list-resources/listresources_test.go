package listresources

import (
	"fmt"
	"testing"

	fake "github.com/gardener/etcd-druid/druidctl/client/fake"
	types "github.com/gardener/etcd-druid/druidctl/cmd/types"
	"github.com/gardener/etcd-druid/druidctl/pkg/log"
	"github.com/gardener/etcd-druid/druidctl/pkg/printer"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestListResources(t *testing.T) {
	streams, _, buf, _ := genericiooptions.NewTestIOStreams()
	configFlags := genericclioptions.NewTestConfigFlags()
	options := &types.Options{
		OutputFormat: printer.OutputTypeNone,
		LogType:      log.LogTypeCharm,
		// ConfigFlags:   configFlags,
		ClientFactory: fake.NewTestFactory(configFlags),
		IOStreams:     streams,
	}
	rootCmd := &cobra.Command{Use: "druidctl"}
	listCmd := NewListResourcesCommand(options)
	rootCmd.AddCommand(listCmd)
	options.AddFlags(rootCmd)

	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"list-resources"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	output := buf.String()
	fmt.Println(output)
	// expectedSubstring := `"etcd": {`
	// if !strings.Contains(output, expectedSubstring) {
	// t.Errorf("Expected output to contain %q, but it did not. Output: %s", expectedSubstring, output)
	// }
}
