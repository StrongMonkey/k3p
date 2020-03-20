package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var (
	LocalChartLocation = ".k3s-chart-data"
	IndexURL           = "https://storage.googleapis.com/k3s-chart-testing-2/index.yaml"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update package.yaml from upsteam",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Reading package list from %v\n", IndexURL)
		indexData, err := httpGet(IndexURL)
		if err != nil {
			handleError(err)
		}

		index := &Index{}
		if err := yaml.Unmarshal(indexData, index); err != nil {
			handleError(err)
		}

		for _, p := range index.Packages {
			chartBasePath := filepath.Join(os.Getenv("HOME"), LocalChartLocation, p.Name)
			fmt.Printf("Removing old data from directory %v\n", chartBasePath)
			if err := os.RemoveAll(chartBasePath); err != nil {
				handleError(err)
			}
			if err := os.MkdirAll(filepath.Join(chartBasePath, p.Name), 0755); err != nil {
				handleError(err)
			}

			fmt.Printf("Reading package data from %v\n", p.URL)
			packageYamlData, err := httpGet(p.URL)
			if err != nil {
				handleError(err)
			}
			packageYaml := &PackageYaml{}
			if err := yaml.Unmarshal(packageYamlData, packageYaml); err != nil {
				handleError(err)
			}
			if err := ioutil.WriteFile(filepath.Join(chartBasePath, "package.yaml"), packageYamlData, 0755); err != nil {
				handleError(err)
			}

			fmt.Printf("Reading chart data from %v\n", packageYaml.Base)
			if err := untar(chartBasePath, packageYaml.Base); err != nil {
				handleError(err)
			}

			fmt.Printf("Applying patches...\n")
			for _, patch := range packageYaml.Patches {
				patchData, err := httpGet(patch.Url)
				if err != nil {
					handleError(err)
				}

				patchFile := filepath.Join(chartBasePath, patch.Name)
				if err := ioutil.WriteFile(patchFile, patchData, 0755); err != nil {
					handleError(err)
				}

				cmd := exec.Command("patch", "--no-backup-if-mismatch", patch.Path, patchFile)
				cmd.Dir = filepath.Join(chartBasePath, "chart")
				if patchResult, err := cmd.CombinedOutput(); err != nil {
					fmt.Println(string(patchResult))
					handleError(err)
				}
			}
		}
		fmt.Println("Reading packages done")
	},
}

func handleError(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func httpGet(url string) ([]byte, error) {
	if strings.HasPrefix(url, "file://") {
		return ioutil.ReadFile(strings.TrimPrefix(url, "file://"))
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func untar(baseDir, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	gzf, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gzf.Close()

	tarReader := tar.NewReader(gzf)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			fallthrough
		case tar.TypeRegA:
			name := header.Name
			contents, err := ioutil.ReadAll(tarReader)
			if err != nil {
				return err
			}
			fileName := filepath.Join(baseDir, name)
			if err := os.MkdirAll(filepath.Dir(fileName), 0755); err != nil {
				return err
			}
			if err := ioutil.WriteFile(filepath.Join(baseDir, name), contents, 0755); err != nil {
				fmt.Println("debug")
				return err
			}
		}
	}

	return nil
}
