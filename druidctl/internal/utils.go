package core

import (
	"context"
	"fmt"

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
