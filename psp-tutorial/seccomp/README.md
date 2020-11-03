# Policy Reference
## Seccomp란
seccomp (secure computing mode의 약자)는 리눅스 커널에서 애플리케이션 샌드박싱 메커니즘을 제공하는 컴퓨터 보안 기능
seccomp은 프로세스가 exit(), sigreturn(), 그리고 이미 열린 파일 디스크립터에 대한 read(), write() 를 제외한 어떠한 시스템 호출도 일으킬 수 없는 안전한 상태로 일방향 변환을 할 수 있게 한다. 만약 다른 시스템 호출을 시도한다면, 커널이 SIGKILL로 프로세스를 종료시킨다. 이러한 의미에서 이것은 시스템의 자원을 가상화하는 것이 아니라 프로세스를 고립시키는 것이라고 할 수 있다.

## Seccomp
seccomp profiles는 PSP annotations로 관리
크게 두가지 annotations가 있다.
- defaultProfileName
- allowedProfileNames

## defaultProfileName
seccomp.security.alpha.kubernetes.io/defaultProfileName - Annotation that specifies the default seccomp profile to apply to containers.

사용 가능한 values는 아래와 같다.
### unconfined
Kube.의 기본값. 다른 value가 주어지지 않는다면, seccomp를 적용하지 않는다.
### runtime/default
default container runtime profile 사용. 자신이 implement한 container runtime version에 해당하는 profile을 사용. 그러나 해당 profile의 경우 많은 권한을 허용하기에, 상업적 사용시에는 application specific한 profile을 설정하기를 권함. (Kubernetes에서는 docker를 사용하기에 docker/default와 동일) Default seccomp profile은 whitelist를 이용하여, 허용하고자 하는 syscall들을 나열한다. 링크에서는 Significant syscalls blocked 리스트도 표로 제공한다.
### docker/default
The Docker default seccomp profile is used. Deprecated as of Kubernetes 1.11. Use runtime/default instead.
### localhost/<path>
Specify a profile as a file on the node located at <seccomp_root>/<path>, where <seccomp_root> is defined via the --seccomp-profile-root flag on the Kubelet.

## allowedProfileNames
seccomp.security.alpha.kubernetes.io/allowedProfileNames - Annotation that specifies which values are allowed for the pod seccomp annotations.

예)
seccomp.security.alpha.kubernetes.io/allowedProfileNames: 'docker/default,runtime/default'
seccomp.security.alpha.kubernetes.io/allowedProfileNames: ‘*’

사용 가능한 values는 위와 같다. *는 모든 profiles를 다 허용한다는 뜻이다.
만일 allowedProfileNames가 정의되어 있지 않다면, defaultProfileName만 사용하겠다는 뜻이다.

# Custom seccomp profiles
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

## seccomp 파일 작성 경로
localhost/<path> - Specify a profile as a file on the node located at <seccomp_root>/<path>, where <seccomp_root> is defined via the --seccomp-profile-root flag on the Kubelet.

kubelet의 seccomp 기본 파일 경로는 --seccomp-profile-root 옵션으로 설정할 수 있다. 옵션으로 설정하지 않는 경우 기본 경로는 default "/var/lib/kubelet/seccomp" 이다.
보통 host에는 default 경로가 존재하지 않기에 폴더를 만들어 줘야 한다. 해당 경로는 localhost의 root 경로가 된다. 이후 위에서 작성한 custom seccomp profile의 이름을 <path>로 명시해주면 된다.

중요 사항
profiles를 적용하고자 하는 모든 host들에 profile 파일을 만들어 줘야 한다.