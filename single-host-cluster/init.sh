#!/bin/bash

# clean up before init
rm $HOME/.kube/config

kubeadm init --config kubeadm-config.yaml

mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

kubectl taint nodes --all node-role.kubernetes.io/master-

# calico
kubectl apply -f calico/tigera-operator.yaml
kubectl apply -f calico/custom-resources.yaml

# kubevirt
kubectl apply -f kubevirt/kubevirt-operator.yaml
kubectl apply -f kubevirt/kubevirt-cr.yaml