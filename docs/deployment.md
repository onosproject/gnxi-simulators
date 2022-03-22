<!--
SPDX-FileCopyrightText: 2022 2020-present Open Networking Foundation <info@opennetworking.org>

SPDX-License-Identifier: Apache-2.0
-->

# Deploying the device simulator

This guide deploys a `device-simulator` through it's [Helm] chart assumes you have a [Kubernetes] cluster running 
with an atomix controller deployed in a namespace.  
`device-simulator` Helm chart is based on Helm 3.0 version, with no need for the Tiller pod to be present.   
If you don't have a cluster running and want to try on your local machine please follow first 
the [Kubernetes] setup steps outlined to [deploy with Helm](https://docs.onosproject.org/developers/deploy_with_helm/).
The following steps assume you have the setup outlined in that page, including the `micro-onos` namespace configured.

Device simulators can be deployed on their own using (from `onos-helm-charts`)
the `device-simulator` chart:

```bash
> helm install -n micro-onos devicesim-1 device-simulator
```
with output along the lines of 
```bash
NAME:   devicesim-1
LAST DEPLOYED: Sun May 12 01:16:41 2019
NAMESPACE: default
STATUS: DEPLOYED
```

The device-simulator chart deploys a single `Pod` containing the device simulator with a `Service`
through which it can be accessed. The device simulator's service can be seen by running the
`kubectl get services` command:

```bash
> kubectl get svc
NAME                              TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
devicesim-1-device-simulator         ClusterIP   10.106.28.52    <none>        10161/TCP        25m
```

### Notify the system about the new simulator.

To notify the system about the newly added simulator or device please go into `onos-cli` and issue the command
```bash
onos topo add device devicesim-1 --address devicesim-1-device-simulator:11161 --type Devicesim --version 1.0.0 --insecure
```

### Installing the chart in a different namespace.

Issue the `helm install` command substituting `micro-onos` with your namespace.
```bash
helm install -n <your_name_space> devicesim-1 device-simulator
```

### Deploying multiple simulators

To deploy multiple simulators, simply install the simulator chart _n_ times
to create _n_ devices, each with a unique name:

```bash
> helm install -n micro-onos devicesim-1 device-simulator
> helm install -n micro-onos devicesim-2 device-simulator
> helm install -n micro-onos devicesim-3 device-simulator
```

### Troubleshoot

If your chart does not install or the pod is not running for some reason and/or you modified values Helm offers two flags to help you
debug your chart: 

* `--dry-run` check the chart without actually installing the pod. 
* `--debug` prints out more information about your chart

```bash
helm install -n micro-onos devicesim-1 --debug --dry-run device-simulator
```

[Helm]: https://helm.sh/
[Kubernetes]: https://kubernetes.io/
