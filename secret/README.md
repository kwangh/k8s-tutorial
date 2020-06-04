## 생성
### create ssh-key secret and pod
kubectl apply -f ssh-key-secret.yaml -f ssh-key-pod.yaml
### create db secret and pod
kubectl apply -f db-secret.yaml -f db-pod.yaml

## 테스트
### ssh-key secret test
kubectl exec -it secret-test-pod -n example -- sh
cat /etc/secret-volume/id-rsa

## db secret test
kubectl exec -it po/test-db-client-pod -n example --sh
kubectl exec -it po/prod-db-client-pod -n example -- sh
cat /etc/secret-volume/password

## 삭제
### ssh-key
kubectl delete -f ssh-key-secret.yaml -f ssh-key-pod.yaml
### db
kubectl delete -f db-secret.yaml -f db-pod.yaml


## kubectl로 직접 secret 생성
### ssh-key 예제
kubectl create secret generic ssh-key-secret --from-file=ssh-privatekey=/root/.ssh/id_rsa --from-file=ssh-publickey=/root/.ssh/id_rsa.pub

### db 예제
kubectl create secret generic prod-db-secret --from-literal=username=produser --from-literal=password=Y4nys7f11


## 그외 cmd
cat ~/.ssh/id_rsa.pub  | base64
echo cHJvZHVzZXI= | base64 --decode
