package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var (
	deleteCustomOptions []string
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a package",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Exact one argument is required")
			os.Exit(1)
		}
		packageName := args[0]

		packagePath := filepath.Join(os.Getenv("HOME"), LocalChartLocation, args[0], "package.yaml")
		if _, err := os.Stat(packagePath); err != nil {
			fmt.Printf("Can't locate package %v. Run `k3p update`.\n", args[0])
		}

		packageYamlData, err := ioutil.ReadFile(packagePath)
		if err != nil {
			handleError(err)
		}
		packageYaml := &PackageYaml{}
		if err := yaml.Unmarshal(packageYamlData, packageYaml); err != nil {
			handleError(err)
		}
		for _, deleteCommand := range packageYaml.PreDeleteCommand {
			args := strings.Fields(deleteCommand)
			if len(args) > 1 {
				c := exec.Command(args[0], args[1:]...)
				fmt.Println(c.Args)
				if output, err := c.CombinedOutput(); err != nil {
					fmt.Println(string(output))
					handleError(err)
				}
			}
		}

		var options []string
		if len(customOptions) > 0 {
			options = append(options, deleteCustomOptions...)
		}
		options = append(append([]string{"delete"}, customOptions...), packageName)
		helmCmd := exec.Command("helm", options...)
		if output, err := helmCmd.CombinedOutput(); err != nil {
			fmt.Println(string(output))
			handleError(err)
		}
	},
}

func init() {
	deleteCmd.Flags().StringArrayVarP(&deleteCustomOptions, "custom-options", "", nil, "custom delete option passed through helm delete")
}
