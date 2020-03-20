package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge CRD configuration for a package",
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

		if packageYaml.CRDManifest != "" {
			fmt.Println("Purging CRDs")
			tmpfile, err := ioutil.TempFile("", fmt.Sprintf("%s-crd-", packageName))
			if err != nil {
				handleError(err)
			}

			if err := ioutil.WriteFile(tmpfile.Name(), []byte(packageYaml.CRDManifest), 0755); err != nil {
				handleError(err)
			}

			kcmd := exec.Command("kubectl", "delete", "-f", tmpfile.Name())
			output, err := kcmd.CombinedOutput()
			if err != nil {
				fmt.Println(string(output))
				handleError(err)
			}
			fmt.Println(string(output))
		}
	},
}
