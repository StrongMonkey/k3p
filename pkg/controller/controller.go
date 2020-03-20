package controller

import (
	"context"
	"fmt"

	"github.com/rancher/k3p/pkg/apis/helm.k3s.io/v1alpha1"
	helmcontroller "github.com/rancher/k3p/pkg/generated/controllers/helm.k3s.io/v1alpha1"
	"github.com/rancher/k3p/types"
	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

var (
	systemNamespace = "kube-system"
)

func Register(ctx context.Context, rContext *types.Context, insecure bool) error {
	apply := rContext.Apply.
		WithCacheTypes(
			rContext.Batch.Batch().V1().Job(),
			rContext.Core.Core().V1().ServiceAccount(),
			rContext.RBAC.Rbac().V1().Role(),
			rContext.RBAC.Rbac().V1().ClusterRole(),
			rContext.RBAC.Rbac().V1().RoleBinding(),
			rContext.RBAC.Rbac().V1().ClusterRoleBinding(),
			).
		WithPatcher(batch.SchemeGroupVersion.WithKind("Job"), func(namespace, name string, pt k8stypes.PatchType, data []byte) (runtime.Object, error) {
			err := rContext.Batch.Batch().V1().Job().Delete(namespace, name, &metav1.DeleteOptions{})
			if err == nil {
				return nil, fmt.Errorf("replace job")
			}
			return nil, nil
		})

	h := handler{
		insecure: insecure,
	}

	helmcontroller.RegisterChartGeneratingHandler(
		ctx,
		rContext.Helm.Helm().V1alpha1().Chart(),
		apply,
		"ChartJobDeployed",
		"k3p-chart",
		h.generate,
		nil,
	)

	return nil
}

type handler struct {
	insecure bool
}

func serviceAccountName(obj *v1alpha1.Chart) string {
	return fmt.Sprintf("%s-sa-install", obj.Name)
}

func roleName(obj *v1alpha1.Chart) string {
	return fmt.Sprintf("%s-role-install", obj.Name)
}

func clusterroleName(obj *v1alpha1.Chart) string {
	return fmt.Sprintf("%s-clusterrole-install", obj.Name)
}

func (h handler) generate(obj *v1alpha1.Chart, status v1alpha1.ChartStatus) ([]runtime.Object, v1alpha1.ChartStatus, error) {
	var result []runtime.Object

	result = append(result, h.generateRbacRoles(obj)...)
	result = append(result, h.generateServiceAccount(obj)...)


}

func (h handler) generateServiceAccount(obj *v1alpha1.Chart) []runtime.Object {
	sa := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: 	serviceAccountName(obj),
			Namespace: obj.Namespace,
		},
	}
	return []runtime.Object{sa}
}

// if controller is set to insecure mode, also create (cluster)roles and (cluster)rolebindings
func (h handler) generateRbacRoles(obj *v1alpha1.Chart) []runtime.Object {
	var result []runtime.Object

	if h.insecure {
		role := obj.Spec.RbacSetting.Roles
		role.Name = fmt.Sprintf("%s-role-install", obj.Name)

		clusterrole := obj.Spec.RbacSetting.ClusterRoles
		clusterrole.Name = fmt.Sprintf("%s-clusterrole-install", obj.Name)

		result = append(result, &role, &clusterrole)
	}

	rolebinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-rolebinding-install"),
			Namespace: obj.Namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     roleName(obj),
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup:  v1.GroupName,
				Name:      serviceAccountName(obj),
				Namespace: obj.Namespace,
			},
		},
	}

	clusterrolebinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-clusterrolebinding-install"),
			Namespace: obj.Namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     clusterroleName(obj),
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup:  v1.GroupName,
				Name:      serviceAccountName(obj),
				Namespace: obj.Namespace,
			},
		},
	}

	result = append(result, rolebinding, clusterrolebinding)

	return result
}

func (h handler) generateValues(obj *v1alpha1.Chart) []string {
	var answerArgs []string
	answerArgs = []string{"--value", "/tmp/values/values.yaml"}
	for k, v := range obj.Spec.ValueOverride {
		answerArgs = append(answerArgs, "--set", fmt.Sprintf("%s=%s", k, v))
	}

	if obj.Spec.PrivateRegistry.Key != "" && obj.Spec.PrivateRegistry.Value != "" {
		answerArgs = append(answerArgs, "--set", fmt.Sprintf("%s=%s", obj.Spec.PrivateRegistry.Key, obj.Spec.PrivateRegistry.Value))
	}

	return answerArgs
}

func (h handler) generateJob(obj *v1alpha1.Chart) error {
	volumeName := "chart-dir"
	mountPath := "/tmp/charts"

	job := batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: obj.Name + "-",
			Namespace:    systemNamespace,
		},
		Spec: batch.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					ServiceAccountName: serviceAccountName(obj),
					Volumes: []v1.Volume{
						{
							Name: volumeName,
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
					},
					Containers: []v1.Container{
						{
							Image: "strongmonkey1992/helm-install",
							VolumeMounts: []v1.VolumeMount{
								{
									Name: volumeName,
									MountPath: mountPath,
								},
							},
							Args: []string{
								"helm",
								"install",
								""
							}
						},
					},
				},
			},
		},
	}
}
