# etcd-druid ctl

>  Decision: Create a kubectl plugin as that is a natural extension to heavily used `kubectl`. 
> Since we are now going to extend kubectl we will be creating the commands and sub-commands using cobra CLI framework.

## etcd-cluster context setting

Inherited from `kubectl`
See https://kubernetes.io/docs/reference/kubectl/generated/kubectl_config/kubectl_config_set-context/
All kubectl plugins get to benefit from inheriting command line arguments like `-A|--all-namespaces`, `-n|--namespace`, `--kubeconfig` etc.

> TODO: Think of a better name for the `kubectl` sub-command. For place holder we are using `druid` but this is not finalised.


## Common CLI flags

Following are the common set of CLI flags that will be applicable to all commands.

| Flag name | Required | Description                                                  |
| --------- | -------- | ------------------------------------------------------------ |
| namespace | No       | If not specified explicitly then it takes the existing namespace (if it is already been targeted via `kubectl config set-context --current --namespace <namespace-name>` or it assumes `default` ns. |

## General management

### Reconcile Etcd resource

```bash
kubectl druid reconcile <etcd-resource-name> --wait-till-ready
# example invocation
# ------------------------------------------------------------
kubectl druid reconcile etcd-main -n shoot-bingo
```

| Flag name       | Required | Description                                                  |
| --------------- | -------- | ------------------------------------------------------------ |
| wait-till-ready | No       | Wait till all changes (if any) done to the `Etcd` resource have successfully reconciled and post reconciliation all the etcd cluster members are `Ready`. |

#### Output

```bash
```

### Resource protection control

> NOTE: This will only have effect if resource protection webhook has been enabled when deploying etcd-druid.

Adds resource protection to all managed components for a given etcd cluster. 

```bash
kubectl druid add-component-protection <etcd-resource-name>
```

Removes resource protection for all managed components for a given etcd cluster.

```bash
kubectl druid remove-component-protection <etcd-resource-name>
```

### Suspend reconciliation

Suspends reconciliation for an etcd cluster.
Some use cases:

* During recovery from quorum loss we suspend etcd cluster reconciliation.
* Transiently change configuration for an etcd cluster. e.g. change the max DB size.

```bash
kubectl druid suspend-reconciliation <etcd-resource-name>
```

To remove suspension of any potential reconciliation one can execute the following command:

```bash
kubectl druid resume-reconciliation <etcd-resource-name>
```

### List all managed resources for an etcd cluster

List resources will list all managed components and also referenced resources like secrets.

```bash
kubectl druid list-resources <etcd-resource-name> --filter=<comma separated types>
# example invocation
# ------------------------------------------------------------
kubectl druid list-resources etcd-main
# Using filter in the following command will only lists resources that are used by etcd-main cluster and specified via filter flag. In the case below these will be Pod, Service, PersistentVolume and Secret resources.
kubectl druid list-resources etcd-main --filter=po,svc,pvc,secret
```

| Flag name | Required | Description                                                  |
| --------- | -------- | ------------------------------------------------------------ |
| filter    | No       | Provides a way to specify command separate resource KINDS that the user is interested to see. If not set then all resources that are managed or referenced by this etcd cluster will be listed. |

### Attach a pre-built debug container 

This command will leverage a non-gardener debug container image that is built with all the utilities to debug an etcd cluster. Since all the etcd pods use `distroless` images an ephemeral container needs to be attached to the distroless container.

> TODO: Enhance the Dockerfile in `etcd-wrapper/ops` and make it available (via an OCI repository) for use as a default debug container image.

```bash
kubectl druid debug etcd-main --member-name=<name of the member> --image=<debug-image-url>
```

| Flag name   | Required | Description                                                  |
| ----------- | -------- | ------------------------------------------------------------ |
| member-name | Yes      | Name of the member which needs to be debugged.               |
| image       | No       | If set then the debug container will use this image. If not specified then it will use the default debug container image. |

> **NOTE:**  Currently member-name is the same as pod-name. For autonomous cluster use case this will change. It is therefore important that we do not assume that a pod-name is always going to be a member-name. From user perspective an etcd cluster member might need to be diagnosed  and for that its sufficient that the user specifies the member name. Internally the implementation needs to get the pod name and attach a debug container to it.

### Hibernate and wake up an etcd cluster

These commands are not supported today till https://github.com/gardener/etcd-druid/issues/922 is implemeted.

To hibernate an etcd cluster execute the following command:

> NOTE: Handling of PVCs (to retain it or delete) is governed by the `Etcd` spec.

```bash
kubectl druid hibernate <etcd-name>
```
To wake up an etcd cluster from hibernation, execute the following command:

```bash
kubectl druid wakeup <etcd-name>
```

### Scale-in/Scale-out of etcd clusters

With https://github.com/gardener/etcd-druid/pull/1070 we re-introduced `Scale` sub-resource for `Etcd` custom resource. This allows anyone to use `kubectl scale`  command to scale-in/out etcd clusters. At the time of writing this doc we were not entirely convinced if we need a sub-command in druidctl for this. If there is a need in future for reasons unknown at this time we can re-consider adding this functionality.

## Monitoring

### Get etcd member status

This sub-command gives you a status for an etcd-cluster which includes the following:

* Output of `etcdctl endpoint status --cluster -wfields` from each member of an etcd cluster.
* Output of `etcdctl member list -wfields` from each member - if these are the same for one or more members then combine the output for easy consumption. In case of a split brain this could be different as seen from individual members.
* All the information that we capture as part of `EtcdMember` resource should also be included.
* *Stretch*: Inspect logs to figure out write/read delays, peer connectivity frequent failures, too many leader elections etc.  and deduce QoS for an etcd cluster.

Consume the output of the above two commands and combine it into easy-to-consume information.

```bash
kubectl druid member-status <etcd-resource-name> --all|--member-name=<name of the member>
```

| Flag name     | Required | Description                                                  |
| ------------- | -------- | ------------------------------------------------------------ |
| --all \| -A   | Optional | If specified then it will give each member status information. |
| --member-name | Optional | If specified it will be give the specified member status information. |

If none of `-A or --member-name` is specified then it shows nothing.

#### Output



### Interactive dashboard	

```bash
kubectl druid dashboard [<etcd-resource-name]> --mode=r|rw (defaulted to 'r')
```

#### Features

*Read only actions:*

* Ability to list all etcd resources
* Show status of each etcd resource
* Drill to etcd cluster members
* Show detailed status of each etcd member
* Get current logs for any etcd member
* If backups are enabled then it can provide the following features:
  * List all available snapshots
  * Download one or more snapshots
  * Inspect snapshots between 2 revisions or duration and generate a report containing the following (list will not be comprehensive but only indicative):
    * Largest prefixes (like du command)
    * Churn rate 
      * rate of prefix size in a given window
      * rate of increase of number of keys in a given window (show the top n key prefixes based on rate of increase)
* Explore etcd db data as a directory for a specific member.

*Mutate actions:*

* All operator tasks

* Etcd resource administration actions (e.g. reconcile, annotate etc.)

  

## Diagnostics



### Diagnose etcd cluster(s)

```bash
kubectl druid diagnose [<etcd-resource-name>] -n=<namespace> -A (across all namespaces) -o[json|yaml]
```



## Operator tasks


