# Kubernetes-podsecuritypolicies
kubernetes psp 번역
kubernetes v1.17

### 참고 사이트
https://kubernetes.io/docs/concepts/policy/pod-security-policy/

https://kubernetes.io/docs/tasks/configure-pod-container/security-context/

---

### 유의 사항
psp 기능은 현재 beta 버전입니다.
테스트 된 코드이며, 안전하며, 기본으로 enable 되어 있습니다.
하지만 기능이 변경 될 수도 있습니다. (기존 releases와 호환이 안될 수도 있습니다.)

## Pod Security Policies란
Pod Security Policies enable fine-grained authorization of pod creation and updates.
A Pod Security Policy (이하 psp) is a cluster-level resource that controls security sensitive aspects of the pod specification. The PodSecurityPolicy objects define a set of conditions that a pod must run with in order to be accepted into the system, as well as defaults for the related fields.

#### 관리자가 관리할 수 있는 항목들
- privileged 컨테이너 여부
- host namespaces 사용
- host network와 host ports
- volume types (https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes)

## PSP 시작
admission controller에 옵션(but recommeded)으로 들어가있다. admission controller를 시작하면 자동으로 PodSecurityPolicies가 적용되는데, policies를 설정해놓지 않고 실행하면 cluster내 모든 pod 생성을 막는다.
pod security policy API (policy/v1beta1/podsecuritypolicy)가 admission controller와는 독립적으로 시작되기 때문에, policies를 추가하고 authorized 할 것을 권장
(API와 PodSecurityPolicies는 별개의 개념. API는 기능셋, PSP는 object 개념)

### admission controller
pod 생성 혹은 변경 시 admission controller가 security context와 pod security policies를 참고하여 허용할지 말지를 결정한다. (https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#podsecuritypolicy)

### admission controller 시작 방법
manifest 방법이 간단하므로 manifest 방법 추천.

#### kubeadm 방법
enable-admission-plugins에 PodSecurityPolicy 추가
```
apiVersion: kubeadm.k8s.io/v1beta2
kind: ClusterConfiguration
kubernetesVersion: "v1.15.3"
controlPlaneEndpoint: "<IP>:6443"
networking:
 serviceSubnet: "10.96.0.0/16"
 podSubnet: "10.244.0.0/16"
apiServer:
 extraArgs:
  enable-admission-plugins: PodSecurityPolicy,NodeRestriction
```

클러스터 구축 후, 아래 psp 적용 필요

```
---
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: privileged
  annotations:
    seccomp.security.alpha.kubernetes.io/allowedProfileNames: "*"
spec:
  privileged: true
  allowPrivilegeEscalation: true
  allowedCapabilities:
  - "*"
  volumes:
  - "*"
  hostNetwork: true
  hostPorts:
  - min: 0
    max: 65535
  hostIPC: true
  hostPID: true
  runAsUser:
    rule: 'RunAsAny'
  seLinux:
    rule: 'RunAsAny'
  supplementalGroups:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: psp:privileged
rules:
- apiGroups: ['policy']
  resources: ['podsecuritypolicies']
  verbs:     ['use']
  resourceNames:
  - privileged
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: default:privileged
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: psp:privileged
subjects:
- kind: Group
  name: system:authenticated
  apiGroup: rbac.authorization.k8s.io
```

#### manifest 방법
cluster 구축 완료한 상태에서
/etc/kubernetes/manifests/kube-apiserver.yaml 파일 수정

PodSecurityPolicy 추가
--enable-admission-plugins=PodSecurityPolicy,NodeRestriction

마찬가지로 위의 psp 추가 필요

## Authorizing Policies
PodSecurityPolicy resource가 생성되고 난 뒤에는, authorize 작업이 필요하다. requesting user or target pod’s service account에 use 권한을 부여해줘야 한다.
일반적으로는 user가 직접 pods을 생성하지 않고 Deployment, ReplicaSet, 혹은 다른 controller manager를 이용하여 생성하기 때문에, service account를 이용하는 방법을 선호.

---
PSP를 사용하려면 kubeadm config에 Admission Controller 옵션으로 PodSecurityPolicy,NodeRestriction을 넣어서 기동시켜줘야 한다.
참고 https://github.com/rancher/k3s/issues/516

그리고는 psp와 clusterrole, clusterrolebinding yaml이 필요하다. 3가지를 적용해주지 않으면 pod들이 정상적으로 뜨지 않는다.
이슈 확인 도중 알게 되었다. https://github.com/kubernetes/minikube/issues/3818

아래 tutorial과 psp 공식 문헌을 참고하여 psp.yaml을 작성하였다.
https://minikube.sigs.k8s.io/docs/tutorials/using_psp/#tutorial

```
---
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: privileged
  annotations:
    seccomp.security.alpha.kubernetes.io/allowedProfileNames: "*"
spec:
  privileged: true
  allowPrivilegeEscalation: true
  allowedCapabilities:
  - "*"
  volumes:
  - "*"
  hostNetwork: true
  hostPorts:
  - min: 0
    max: 65535
  hostIPC: true
  hostPID: true
  runAsUser:
    rule: 'RunAsAny'
  seLinux:
    rule: 'RunAsAny'
  supplementalGroups:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
---
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: restricted
  annotations:
    seccomp.security.alpha.kubernetes.io/allowedProfileNames: 'docker/default,runtime/default'
    apparmor.security.beta.kubernetes.io/allowedProfileNames: 'runtime/default'
    seccomp.security.alpha.kubernetes.io/defaultProfileName:  'runtime/default'
    apparmor.security.beta.kubernetes.io/defaultProfileName:  'runtime/default'
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  hostNetwork: false
  hostIPC: false
  hostPID: false
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  supplementalGroups:
    rule: 'MustRunAs'
    ranges:
      # Forbid adding the root group.
      - min: 1
        max: 65535
  fsGroup:
    rule: 'MustRunAs'
    ranges:
      # Forbid adding the root group.
      - min: 1
        max: 65535
  readOnlyRootFilesystem: false
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: psp:privileged
rules:
- apiGroups: ['policy']
  resources: ['podsecuritypolicies']
  verbs:     ['use']
  resourceNames:
  - privileged
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: psp:restricted
rules:
- apiGroups: ['policy']
  resources: ['podsecuritypolicies']
  verbs:     ['use']
  resourceNames:
  - restricted
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: default:restricted
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: psp:restricted
subjects:
- kind: Group
  name: system:authenticated
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: default:privileged
  namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: psp:privileged
subjects:
- kind: Group
  name: system:masters
  apiGroup: rbac.authorization.k8s.io
- kind: Group
  name: system:nodes
  apiGroup: rbac.authorization.k8s.io
- kind: Group
  name: system:serviceaccounts:kube-system
  apiGroup: rbac.authorization.k8s.io
```

위의 yaml을 간략하게 설명하면,
kube-system에서 필요로 하는 clusterrolebinding으로 system:masters, system:nodes, system:serviceaccounts:kube-system에 privileged 권한을 주었다.
- system:masters는 cluster-admin(Default clusterrole)에 bind 되어있는데, 이는 super user가 resource에 하는 모든 action에 대한 권한 제공.
- system:nodes는 system:node(Default clusterrole)에 bind 되어 있으며, kubelet component가 필요로 하는 resource들에 권한 제공.
- system:serviceaccounts:kube-system은 kube-system namespace의 모든 service accounts에 권한 제공.

권한관리 개념은 조대협 블로그에 한글 설명이 있으니 참고하시면 됩니다. https://bcho.tistory.com/1272?category=731548
