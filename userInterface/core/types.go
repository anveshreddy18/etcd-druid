package core

import (
	"context"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type druidEtcdClient interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*druidv1alpha1.Etcd, error)
	Update(ctx context.Context, etcd *druidv1alpha1.Etcd, opts metav1.UpdateOptions) (*druidv1alpha1.Etcd, error)
}
