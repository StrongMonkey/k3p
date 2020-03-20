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

var (
	customOptions []string
	profile       string
	updateCrdOnly bool
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install packages",
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

		if packageYaml.CRDManifest != "" && updateCrdOnly {
			fmt.Println("Upgrading CRDs")
			tmpfile, err := ioutil.TempFile("", fmt.Sprintf("%s-crd-", packageName))
			if err != nil {
				handleError(err)
			}

			if err := ioutil.WriteFile(tmpfile.Name(), []byte(packageYaml.CRDManifest), 0755); err != nil {
				handleError(err)
			}

			kcmd := exec.Command("kubectl", "apply", "-f", tmpfile.Name())
			if output, err := kcmd.CombinedOutput(); err != nil {
				fmt.Println(string(output))
				handleError(err)
			}
			return
		}

		fmt.Println("Install helm releases")
		// run helm install
		var options []string
		tmpfileValue, err := ioutil.TempFile("", fmt.Sprintf("%s-value-", packageName))
		if err != nil {
			handleError(err)
		}
		for name, prof := range packageYaml.ProfileOptions {
			if profile != "" {
				if profile == name {
					if err := ioutil.WriteFile(tmpfileValue.Name(), []byte(prof.ValueYaml), 0755); err != nil {
						return
					}
				}
				continue
			} else if prof.Default {
				if err := ioutil.WriteFile(tmpfileValue.Name(), []byte(prof.ValueYaml), 0755); err != nil {
					return
				}
			}
		}
		options = []string{"--values", tmpfileValue.Name()}

		if packageYaml.PrivateRegistry.Key != "" && packageYaml.PrivateRegistry.Value != "" {
			options = append(options, "--set", fmt.Sprintf("%s=%s", packageYaml.PrivateRegistry.Key, packageYaml.PrivateRegistry.Value))
		}

		if len(customOptions) > 0 {
			options = append(options, customOptions...)
		}

		chartDir := filepath.Join(os.Getenv("HOME"), LocalChartLocation, packageName, "chart")
		helmArgs := append([]string{"upgrade"}, append(options, "--install", packageName, chartDir)...)
		helmCmd := exec.Command("helm", helmArgs...)
		output, err := helmCmd.CombinedOutput()
		if err != nil {
			fmt.Println(string(output))
			handleError(err)
		}
		fmt.Println(string(output))
	},
}

func init() {
	installCmd.Flags().BoolVarP(&updateCrdOnly, "update-crd-only", "", false, "only update the crd")
	installCmd.Flags().StringVarP(&profile, "profile", "p", "", "profile is a set of answer values for a helm chart")
	installCmd.Flags().StringArrayVarP(&customOptions, "custom-options", "", nil, "pass custom helm options")
}
