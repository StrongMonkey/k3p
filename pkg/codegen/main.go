package main

import (
    "github.com/rancher/k3p/pkg/apis/helm.k3s.io/v1alpha1"
    controllergen "github.com/rancher/wrangler/pkg/controller-gen"
    "github.com/rancher/wrangler/pkg/controller-gen/args"

)

var (
    basePackage = "github.com/rancher/k3p/types"
)

func main() {
    controllergen.Run(args.Options{
        OutputPackage: "github.com/rancher/k3p/pkg/generated",
        Boilerplate:   "scripts/boilerplate.go.txt",
        Groups: map[string]args.Group{
            "helm.k3s.io": {
                Types: []interface{}{
                    v1alpha1.Chart{},
                },
                GenerateTypes: true,
            },
        },
    })
}
