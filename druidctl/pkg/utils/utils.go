package utils

import (
	"context"
	"fmt"
	"time"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	client "github.com/gardener/etcd-druid/druidctl/client"
)

func GetEtcdList(ctx context.Context, cl client.EtcdClientInterface, name, namespace string, allNamespaces bool) (*druidv1alpha1.EtcdList, error) {
	etcdList := &druidv1alpha1.EtcdList{}
	var err error
	if allNamespaces {
		etcdList, err = cl.ListEtcds(ctx, "")
		if err != nil {
			return nil, fmt.Errorf("unable to list etcd objects: %w", err)
		}
	} else {
		etcd, err := cl.GetEtcd(ctx, namespace, name)
		if err != nil {
			return nil, fmt.Errorf("unable to get etcd object: %w", err)
		}
		etcdList.Items = append(etcdList.Items, *etcd)
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
