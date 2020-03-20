package types

import (
	"context"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/core"

	"github.com/rancher/k3p/pkg/generated/controllers/helm.k3s.io"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/batch"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/rbac"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/start"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type contextKey struct{}

type Config interface {
	Get(section, name string) string
}

type Context struct {
	Namespace string

	Core  *core.Factory
	Batch *batch.Factory
	RBAC  *rbac.Factory
	Helm  *helm.Factory
	K8s   kubernetes.Interface
	Apply apply.Apply
}

func From(ctx context.Context) *Context {
	return ctx.Value(contextKey{}).(*Context)
}

func NewContext(namespace string, config *rest.Config) *Context {
	context := &Context{
		Namespace: namespace,
		Core:      core.NewFactoryFromConfigOrDie(config),
		RBAC:      rbac.NewFactoryFromConfigOrDie(config),
		K8s:       kubernetes.NewForConfigOrDie(config),
		Helm:      helm.NewFactoryFromConfigOrDie(config),
		Batch:     batch.NewFactoryFromConfigOrDie(config),
	}

	context.Apply = apply.New(context.K8s.Discovery(), apply.NewClientFactory(config)).WithRateLimiting(20.0)
	return context
}

func (c *Context) Start(ctx context.Context) error {
	return start.All(ctx, 5,
	    c.RBAC,
		c.Core,
		c.Batch,
		c.Helm)
}

func BuildContext(ctx context.Context, namespace string, config *rest.Config) (context.Context, *Context) {
	c := NewContext(namespace, config)
	return context.WithValue(ctx, contextKey{}, c), c
}

func Register(f func(context.Context, *Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return f(ctx, From(ctx))
	}
}
