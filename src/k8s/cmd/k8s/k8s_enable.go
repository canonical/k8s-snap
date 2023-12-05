package k8s

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/canonical/k8s/pkg/k8s/component"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	var enableCmdOpts struct {
		sets []string
		file string
	}

	enableCmd := &cobra.Command{
		Use:       "enable [component]",
		Short:     "Enable a specific component in the cluster",
		Long:      "Enable one of the specific components: cni, dns, gateway or ingress.",
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{"cni", "dns", "gateway", "ingress"},
		RunE:      runEnableCmd(&enableCmdOpts),
	}

	enableCmd.Flags().StringSliceVar(&enableCmdOpts.sets, "set", []string{}, "Set values for the chart (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	enableCmd.Flags().StringVarP(&enableCmdOpts.file, "file", "f", "", "Optional YAML file with the configuration options for the component")
	rootCmd.AddCommand(enableCmd)
}

func runEnableCmd(opts *struct {
	sets []string
	file string
}) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if opts.file != "" {
			if _, err := os.Stat(opts.file); os.IsNotExist(err) {
				return fmt.Errorf("file does not exist: %s", opts.file)
			}
		}

		values := parseSets(opts.sets)
		if values == nil {
			return errors.New("invalid format for --set, expected key=value")
		}
		if err := component.EnableComponent(args[0], values, opts.file); err != nil {
			return err
		}

		logrus.Infof("Component %s enabled", args[0])
		return nil

	}
}

// TODO: (mateoflorido) Replicate nesting set mechanism.
func parseSets(sets []string) map[string]interface{} {
	values := make(map[string]interface{})
	for _, set := range sets {
		pair := strings.SplitN(set, "=", 2)
		if len(pair) != 2 {
			return nil
		}
		values[pair[0]] = pair[1]
	}
	return values
}
