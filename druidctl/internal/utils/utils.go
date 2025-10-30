package utils

import (
	"context"
	"fmt"
	"sync"
	"time"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	"github.com/gardener/etcd-druid/druidctl/internal/client"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	configFlags     *genericclioptions.ConfigFlags
	configFlagsOnce sync.Once
)

// GetConfigFlags returns a singleton *ConfigFlags for kubeconfig and context handling.
func GetConfigFlags() *genericclioptions.ConfigFlags {
	configFlagsOnce.Do(func() {
		configFlags = genericclioptions.NewConfigFlags(true)
	})
	return configFlags
}

func GetEtcdList(ctx context.Context, cl client.EtcdClientInterface, etcdRefList []types.NamespacedName, allNamespaces bool) (*druidv1alpha1.EtcdList, error) {
	etcdList := &druidv1alpha1.EtcdList{}
	var err error
	if allNamespaces {
		etcdList, err = cl.ListEtcds(ctx, "")
		if err != nil {
			return nil, fmt.Errorf("unable to list etcd objects: %w", err)
		}
	} else {
		for _, ref := range etcdRefList {
			if ref.Name == "*" {
				nsEtcdList, err := cl.ListEtcds(ctx, ref.Namespace)
				if err != nil {
					return nil, fmt.Errorf("unable to list etcd objects in namespace %s: %w", ref.Namespace, err)
				}
				etcdList.Items = append(etcdList.Items, nsEtcdList.Items...)
				continue
			}
			etcd, err := cl.GetEtcd(ctx, ref.Namespace, ref.Name)
			if err != nil {
				return nil, fmt.Errorf("unable to get etcd object: %w", err)
			}
			etcdList.Items = append(etcdList.Items, *etcd)
		}
	}
	return etcdList, nil
}

func ShortDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	days := int(d.Hours()) / 24
	return fmt.Sprintf("%dd", days)
}
