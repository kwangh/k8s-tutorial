# custom-seccomp-profile
custom seccomp profile

## Tutorial

### start with the configmap & daemonset
configmap contains profile data
daemonset's for installing profile data on worker nodes
```
kubectl apply -f seccomp-installer.yaml
```

check configmap lists
```
kubectl get cm -A
```

### results
```
NAME              DATA   AGE
seccomp-profile   1      16m
```

check daemonset pods
```
kubectl -n kube-system get pods -l security=seccomp
```

### results
```
NAME            READY   STATUS    RESTARTS   AGE
seccomp-ss5l9   1/1     Running   0          41m
seccomp-wwf6n   1/1     Running   0          41m
```

### now create a pod with app container
```
kubectl apply -f pod.yaml
```

to use seccomp profile in pod, annotations are needed. container example are also included
```
annotations:
    seccomp.security.alpha.kubernetes.io/pod: "localhost/my-profile.json"
    # container.seccomp.security.alpha.kubernetes.io/<myapp-container>: "localhost/my-profile.json"
```

check pod lists
```
kubectl get pod -A
```

### results
```
NAMESPACE     NAME                                      READY   STATUS    RESTARTS   AGE
default       seccomp-app                               1/1     Running   0          26m
```

Get a shell to the running Container

```
kubectl exec -it seccomp-app -- /bin/bash
```


## Reference

Seccomp & configmap
- https://gardener.cloud/050-tutorials/content/howto/secure-seccomp/
- https://kubernetes.io/docs/concepts/policy/pod-security-policy/
- https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/

Container
- https://kubernetes.io/docs/tasks/debug-application-cluster/get-shell-running-container/
- https://kubernetes.io/ko/docs/concepts/workloads/pods/init-containers/
