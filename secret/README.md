## 생성
kubectl apply -f secret.yaml
kubectl apply -f pod.yaml

## 테스트
kubectl exec -it secret-test-pod -n example -- sh
cat /etc/secret-volume/id-rsa

## 삭제
kubectl delete -f pod.yaml -f secret.yaml


## kubectl로 직접 secret 생성
kubectl create secret generic ssh-key-secret --from-file=ssh-privatekey=/root/.ssh/id_rsa --from-file=ssh-publickey=/root/.ssh/id_rsa.pub

## 그외 cmd
cat ~/.ssh/id_rsa.pub  | base64
