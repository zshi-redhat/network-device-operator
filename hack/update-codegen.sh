#!/bin/bash
unset GOFLAGS
unset GO111MODULE

mkdir -p api/netdev
cd api/netdev
ln -s ../v1alpha1 ./v1alpha1
cd ../..

CODEGEN_PKG="./vendor/k8s.io/code-generator"

bash ${CODEGEN_PKG}/generate-groups.sh client,lister,informer \
      github.com/zshi-redhat/network-device-operator/pkg/client \
      github.com/zshi-redhat/network-device-operator/api \
      netdev:v1alpha1 \
      --go-header-file hack/boilerplate.go.txt

rm -rf pkg/client/*
mv github.com/zshi-redhat/network-device-operator/pkg/client/* ./pkg/client/
sed -i "s|github.com/zshi-redhat/network-device-operator/api/netdev/v1alpha1|github.com/zshi-redhat/network-device-operator/api/v1alpha1|g" $(find pkg/client -type f)
rm -rf api/netdev
