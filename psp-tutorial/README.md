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


## Policy Reference
### Seccomp란
seccomp (secure computing mode의 약자)는 리눅스 커널에서 애플리케이션 샌드박싱 메커니즘을 제공하는 컴퓨터 보안 기능
seccomp은 프로세스가 exit(), sigreturn(), 그리고 이미 열린 파일 디스크립터에 대한 read(), write() 를 제외한 어떠한 시스템 호출도 일으킬 수 없는 안전한 상태로 일방향 변환을 할 수 있게 한다. 만약 다른 시스템 호출을 시도한다면, 커널이 SIGKILL로 프로세스를 종료시킨다. 이러한 의미에서 이것은 시스템의 자원을 가상화하는 것이 아니라 프로세스를 고립시키는 것이라고 할 수 있다.

### Seccomp
seccomp profiles는 PSP annotations로 관리
크게 두가지 annotations가 있다.
- defaultProfileName
- allowedProfileNames

### defaultProfileName
seccomp.security.alpha.kubernetes.io/defaultProfileName - Annotation that specifies the default seccomp profile to apply to containers.

사용 가능한 values는 아래와 같다.
#### unconfined
Kube.의 기본값. 다른 value가 주어지지 않는다면, seccomp를 적용하지 않는다.
#### runtime/default
default container runtime profile 사용. 자신이 implement한 container runtime version에 해당하는 profile을 사용. 그러나 해당 profile의 경우 많은 권한을 허용하기에, 상업적 사용시에는 application specific한 profile을 설정하기를 권함. (Kubernetes에서는 docker를 사용하기에 docker/default와 동일) Default seccomp profile은 whitelist를 이용하여, 허용하고자 하는 syscall들을 나열한다. 링크에서는 Significant syscalls blocked 리스트도 표로 제공한다.
#### docker/default
The Docker default seccomp profile is used. Deprecated as of Kubernetes 1.11. Use runtime/default instead.
#### localhost/<path>
Specify a profile as a file on the node located at <seccomp_root>/<path>, where <seccomp_root> is defined via the --seccomp-profile-root flag on the Kubelet.

### allowedProfileNames
seccomp.security.alpha.kubernetes.io/allowedProfileNames - Annotation that specifies which values are allowed for the pod seccomp annotations.

예)
seccomp.security.alpha.kubernetes.io/allowedProfileNames: 'docker/default,runtime/default'
seccomp.security.alpha.kubernetes.io/allowedProfileNames: ‘*’

사용 가능한 values는 위와 같다. *는 모든 profiles를 다 허용한다는 뜻이다.
만일 allowedProfileNames가 정의되어 있지 않다면, defaultProfileName만 사용하겠다는 뜻이다.

## Custom seccomp profiles
custom profile.json을 작성 하기위해, docker의 deafult profile을 참고하면
https://github.com/moby/moby/blob/master/profiles/seccomp/default.json
https://docs.docker.com/engine/security/seccomp/

```
{
	"defaultAction": "SCMP_ACT_ERRNO",
	"archMap": [
		{
			"architecture": "SCMP_ARCH_X86_64",
			"subArchitectures": [
				"SCMP_ARCH_X86",
				"SCMP_ARCH_X32"
			]
		},
		… 이하 생략
	],
	"syscalls": [
		{
			"names": [
				"accept",
… 생략 ...
				"epoll_create",
				"write"
			],
			"action": "SCMP_ACT_ALLOW",
			"args": [],
			"comment": "",
			"includes": {},
			"excludes": {}
		},
 … 이하 생략 ...
}
```

기본적으로 docker는 화이트리스트를 작성한다. 즉, 허용하고자 하는 syscalls를 제외한 모든 syscalls를 막는다. (만약 블랙리스트를 작성하고자 한다면, defaultAction과 action을 반대로 작성하면 된다.)
위에서 예의 defaultAction의 SCMP_ACT_ERRNO를 통해 모든 syscalls를 막는다고 명시한다.
그리고 허용하는 syscalls는 action의 SCMP_ACT_ALLOW를 통해 허용해준다.
architecture는 허용하고자 하는 syscalls가 동작하는 container runtime의 architecture를 정한다.

허용하고자 하는 syscalls가 명확하지 않은 경우, defaultAction에 SCMT_ACT_LOG을 사용하여, 어떤 syscalls들이 사용되고 있는지 log로 떨궈보는 작업이 도움된다.

### seccomp 파일 작성 경로
localhost/<path> - Specify a profile as a file on the node located at <seccomp_root>/<path>, where <seccomp_root> is defined via the --seccomp-profile-root flag on the Kubelet.

kubelet의 seccomp 기본 파일 경로는 --seccomp-profile-root 옵션으로 설정할 수 있다. 옵션으로 설정하지 않는 경우 기본 경로는 default "/var/lib/kubelet/seccomp" 이다.
보통 host에는 default 경로가 존재하지 않기에 폴더를 만들어 줘야 한다. 해당 경로는 localhost의 root 경로가 된다. 이후 위에서 작성한 custom seccomp profile의 이름을 <path>로 명시해주면 된다.

중요 사항
profiles를 적용하고자 하는 모든 host들에 profile 파일을 만들어 줘야 한다.

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

---
## Policy Reference

### Privileged
pod에 있는 container에 privileged 권한을 할당. 기본적으로는 container는 호스트의 devices에 접근 권한이 없으나, "privileged" container는 host의 모든 devices에 접근 권한을 갖는다. 이는 호스트에서 돌고 있는 프로세스와 거의 동일한 접근 권한을 container가 갖게 되는 것이다. Linux capabilities를 사용하고자 하는 containers에게 유용하다. 예를 들어 network stack을 조작하거나 devices에 접근하고자 하는 경우.

### Host namespaces
- HostPID - pod containers가 host process ID namespace를 공유하도록 변경 가능. ptrace와 같이 사용하면 container 외부로 privilege excalation이 가능하다. 기본으로 ptrace는 막혀있다.
- HostIPC - pod이 host IPC namespace를 공유
- HostNetwork - pod이 node의 network namespace를 사용 가능하도록 변경 가능. pod으로 하여금 loopback device를 사용 가능, localhost를 listen 하는 서비스 사용 가능, 같은 노드의 다른 pod의 network activity를 확인(snoop) 가능
- HostPorts - host network namespace의 사용가능한 ports를 제공. HostPortRange를 이용해서, min(명시된 숫자 포함), max(명시된 숫자 포함) 정의 가능. 기본으로는 host ports를 허용하지 않음

