//go:generate go run pkg/codegen/cleanup/main.go
//go:generate /bin/rm -rf pkg/generated
//go:generate go run pkg/codegen/main.go

package main

import (
	"flag"
	"fmt"
	"github.com/rancher/k3p/pkg/controller"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/rancher/k3p/types"
	"github.com/rancher/wrangler/pkg/kubeconfig"
	"github.com/rancher/wrangler/pkg/leader"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/util/runtime"
)

var (
	Version         = "v0.0.0-dev"
	GitCommit       = "HEAD"
	systemNamespace = "k3s-package"
)

func main() {
	app := cli.NewApp()
	app.Name = "k3p"
	app.Version = fmt.Sprintf("%s (%s)", Version, GitCommit)
	app.Usage = "testy needs help!"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "kubeconfig",
			EnvVar: "KUBECONFIG",
			Value:  "${HOME}/.kube/config",
		},
		cli.BoolFlag{
			Name:        "rbac-insecure",
			Usage:       "if true, user will not need to create role/clusterrole first",
			EnvVar:      "PRIVILEGDED",
		},
	}
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) {
	flag.Parse()
	logrus.Info("Starting controller")

	go func() {
		err := http.ListenAndServe("127.0.0.1:6061", nil)
		if err != nil {
			logrus.Errorf("Failed to launch pprof on port 6061: %v", err)
		}
	}()

	loader := kubeconfig.GetInteractiveClientConfig(c.String("kubeconfig"))
	restConfig, err := loader.ClientConfig()
	if err != nil {
		panic(err)
	}

	ctx, k3pContext := types.BuildContext(context.Background(), systemNamespace, restConfig)

	leader.RunOrDie(ctx, systemNamespace, "k3p", k3pContext.K8s, func(ctx context.Context) {
		if err := controller.Register(ctx, k3pContext, c.Bool("insecure")); err != nil {
			panic(err)
		}
		runtime.Must(k3pContext.Start(ctx))
		<-ctx.Done()
	})
}
