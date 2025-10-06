package core

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gardener/etcd-druid/druidctl/cli/types"
	"github.com/gardener/etcd-druid/druidctl/pkg/output"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type resourceKey struct {
	Group    string
	Version  string
	Resource string
	Kind     string
}

type resourceRef struct {
	Namespace string
	Name      string
	Age       time.Duration
	Labels    map[string]string
	OwnerRefs []ownerRefLite
}

type ownerRefLite struct {
	APIVersion string
	Kind       string
	Name       string
}

type etcdRef struct {
	Name      string
	Namespace string
}

type etcdResourceSummary struct {
	Etcd  etcdRef
	Items map[resourceKey][]resourceRef
}

func ListManagedResources(ctx context.Context, listResourcesCommandCtx *types.ListResourcesCommandContext) error {
	out := listResourcesCommandCtx.Output

	etcdClient := listResourcesCommandCtx.EtcdClient
	genClient := listResourcesCommandCtx.GenericClient

	tokens := parseFilter(listResourcesCommandCtx.Filter)
	if len(tokens) == 0 || (len(tokens) == 1 && tokens[0] == "all") {
		tokens = defaultResourceTokens()
	}
	resolver, err := NewAPIResourceResolver(genClient.Discovery())
	if err != nil {
		return fmt.Errorf("failed to initialize resource resolver: %w", err)
	}
	metas, err := resolver.Resolve(tokens)
	if err != nil {
		return err
	}

	// Identify etcds to operate on
	etcdList, err := GetEtcdList(ctx, etcdClient, listResourcesCommandCtx.ResourceName, listResourcesCommandCtx.Namespace, listResourcesCommandCtx.AllNamespaces)
	if err != nil {
		return err
	}
	if len(etcdList.Items) == 0 {
		if listResourcesCommandCtx.AllNamespaces {
			out.Info("No Etcd resources found across all namespaces")
		} else {
			return fmt.Errorf("etcd %q not found in namespace %q", listResourcesCommandCtx.ResourceName, listResourcesCommandCtx.Namespace)
		}
		return nil
	}

	// Collect results per etcd
	results := make([]etcdResourceSummary, 0, len(etcdList.Items))
	for _, e := range etcdList.Items {
		summary := etcdResourceSummary{
			Etcd:  etcdRef{Name: e.Name, Namespace: e.Namespace},
			Items: map[resourceKey][]resourceRef{},
		}

		selector := fmt.Sprintf("app.kubernetes.io/part-of=%s", e.Name)
		for _, m := range metas {
			// Skip cluster-scoped if not intended; most curated resources are namespaced.
			nsIf := ""
			if m.Namespaced {
				nsIf = e.Namespace
			}
			ulist, err := genClient.Dynamic().Resource(m.GVR).Namespace(nsIf).List(ctx, metav1.ListOptions{LabelSelector: selector})
			if err != nil {
				out.Warning("Failed to list ", m.GVR.Resource, " for etcd ", e.Name, ": ", err.Error())
				continue
			}
			if len(ulist.Items) == 0 {
				continue
			}
			rk := resourceKey{Group: m.GVR.Group, Version: m.GVR.Version, Resource: m.GVR.Resource, Kind: m.Kind}
			for _, item := range ulist.Items {
				summary.Items[rk] = append(summary.Items[rk], toResourceRef(&item))
			}
		}
		// Sort within each resource kind by namespace/name for determinism
		for k := range summary.Items {
			sort.Slice(summary.Items[k], func(i, j int) bool {
				ai, aj := summary.Items[k][i], summary.Items[k][j]
				if ai.Namespace == aj.Namespace {
					return ai.Name < aj.Name
				}
				return ai.Namespace < aj.Namespace
			})
		}
		results = append(results, summary)
	}

	// Sort etcds by namespace/name
	sort.Slice(results, func(i, j int) bool {
		if results[i].Etcd.Namespace == results[j].Etcd.Namespace {
			return results[i].Etcd.Name < results[j].Etcd.Name
		}
		return results[i].Etcd.Namespace < results[j].Etcd.Namespace
	})

	renderListResources(out, results)
	return nil
}

// parseFilter splits and normalizes the filter string
func parseFilter(filter string) []string {
	if strings.TrimSpace(filter) == "" {
		return nil
	}
	parts := strings.Split(filter, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, p := range parts {
		t := strings.ToLower(strings.TrimSpace(p))
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

// defaultResourceTokens returns the curated default set for "all"
func defaultResourceTokens() []string {
	return []string{"po", "sts", "svc", "cm", "secret", "pvc", "lease", "pdb", "role", "rolebinding", "sa"}
}

func toResourceRef(u *unstructured.Unstructured) resourceRef {
	var owners []ownerRefLite
	for _, o := range u.GetOwnerReferences() {
		owners = append(owners, ownerRefLite{APIVersion: o.APIVersion, Kind: o.Kind, Name: o.Name})
	}
	age := time.Since(u.GetCreationTimestamp().Time)
	return resourceRef{
		Namespace: u.GetNamespace(),
		Name:      u.GetName(),
		Age:       age,
		Labels:    u.GetLabels(),
		OwnerRefs: owners,
	}
}

// renderListResources prints results in a grouped, neat format using the Output service.
func renderListResources(out output.Service, results []etcdResourceSummary) {
	for _, s := range results {
		out.Header(fmt.Sprintf("Etcd %s/%s", s.Etcd.Namespace, s.Etcd.Name))
		if len(s.Items) == 0 {
			out.Info("No resources found for selected filters")
			continue
		}
		// Order resource kinds consistently
		keys := make([]resourceKey, 0, len(s.Items))
		for k := range s.Items {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			if keys[i].Kind == keys[j].Kind {
				return keys[i].Resource < keys[j].Resource
			}
			return keys[i].Kind < keys[j].Kind
		})

		for _, k := range keys {
			list := s.Items[k]
			out.RawHeader(fmt.Sprintf("%s (%s.%s/%s): %d", k.Kind, k.Resource, k.Group, k.Version, len(list)))
			for _, r := range list {
				age := shortDuration(r.Age)
				ns := r.Namespace
				if ns == "" {
					ns = "-"
				}
				out.Info(fmt.Sprintf("%s/%s (age %s)", ns, r.Name, age))
			}
		}
	}
}
