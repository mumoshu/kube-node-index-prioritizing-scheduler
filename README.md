# kube-node-index-prioritizing-scheduler

A Kubernetes scheduler that uses sorted nodes' positions as priorities. Use in combination with resource request/limit to implement low-latency highly available front proxy cluster.

Based on the awesome [k8s-scheduler-extender-example](https://github.com/everpeace/k8s-scheduler-extender-example) example by @everpeace.

## Use-cases

1. Accelerate `Envoy` front proxies or Istio ingress gateways. Or more generally, use with Kubernetes `Service` with `type: NodePort` and `externalTrafficPolicy: Local`.

`externalTrafficPolicy: Local` eliminates the extra network hop between the NodePort and the service-backend pods.

But it results in a lower availability due to that a Node with all the backend pods failed results in a complete failure of the NodePort.

With `kube-node-index-prioritizing-scheduler`, you can do you best to keep scheduling `N` pods per node so that `N - 1` pods failure in a single Node doesn't affect availability.

## How to

### 1. Buid a Docker image

```
$ IMAGE=YOUR_ORG/YOUR_IMAGE:YOUR_TAG make build push
```

### 2. Deploy `my-scheduler` onto the `kube-system` namespace

Please see ConfigMap in [extender.yaml](extender.yaml) for scheduler policy json which includes scheduler extender config.

```
$ IMAGE=YOUR_ORG/YOUR_IMAGE:YOUR_TAG make deploy
```

For ease of observation, start streaming logs from the extender:

```console
$ kubectl -n kube-system logs deploy/my-scheduler -c my-scheduler-extender-ctr -f
[  warn ] 2018/11/07 08:41:40 main.go:84: LOG_LEVEL="" is empty or invalid, fallling back to "INFO".
[  info ] 2018/11/07 08:41:40 main.go:98: Log level was set to INFO
[  info ] 2018/11/07 08:41:40 main.go:116: server starting on the port :80
```

```console
$ kubectl -n kube-system logs deploy/my-scheduler -c my-scheduler-ctr -f
...
```

Open up an another termianl and proceed.

### 3. schedule test pod

You will see `test-deploy` pods will be scheduled by `my-scheduler` preferring the nodes with the names that are earlier in the alphabetical order:

```
$ make tryit

$ kubectl logs -f deploy/my-scheduler -c my-scheduler-ctr
...snip...
I1112 08:44:16.854457       1 generic_scheduler.go:676] Host ip-10-0-1-72.ap-northeast-1.compute.internal => Score 107
I1112 08:44:16.854477       1 generic_scheduler.go:676] Host ip-10-0-1-9.ap-northeast-1.compute.internal => Score 109
I1112 08:44:16.854489       1 generic_scheduler.go:676] Host ip-10-0-1-236.ap-northeast-1.compute.internal => Score 74
I1112 08:44:16.854494       1 generic_scheduler.go:676] Host ip-10-0-1-59.ap-northeast-1.compute.internal => Score 98
...snip...

$ kubectl describe pod test-deploy-...
Name:         test-pod
...
Events:
  Type    Reason                 Age   From               Message
  ----    ------                 ----  ----               -------
  Normal  Scheduled              25s   my-scheduler       Successfully assigned test-pod to minikube
  Normal  SuccessfulMountVolume  25s   kubelet, minikube  MountVolume.SetUp succeeded for volume "default-token-wrk5s"
  Normal  Pulling                24s   kubelet, minikube  pulling image "nginx"
  Normal  Pulled                 8s    kubelet, minikube  Successfully pulled image "nginx"
  Normal  Created                8s    kubelet, minikube  Created container
  Normal  Started                8s    kubelet, minikube  Started container
```
