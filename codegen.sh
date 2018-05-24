#!/bin/bash

set -x

GOPATH=$(go env GOPATH)
PACKAGE_NAME=sigs.k8s.io/cluster-api
REPO_ROOT="$GOPATH/src/$PACKAGE_NAME"
DOCKER_REPO_ROOT="/go/src/$PACKAGE_NAME"
DOCKER_CODEGEN_PKG="/go/src/k8s.io/code-generator"

pushd $REPO_ROOT

# Generate openapi
docker run --rm -ti -u $(id -u):$(id -g) \
    -v "$REPO_ROOT":"$DOCKER_REPO_ROOT" \
    -w "$DOCKER_REPO_ROOT" \
    appscode/gengo:release-1.10 openapi-gen \
    --v 1 --logtostderr \
    --go-header-file "boilerplate.go.txt" \
    --input-dirs "$PACKAGE_NAME/pkg/apis/cluster/v1alpha1,k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/api/resource,k8s.io/apimachinery/pkg/runtime,k8s.io/apimachinery/pkg/util/intstr,k8s.io/apimachinery/pkg/version,k8s.io/api/core/v1" \
    --output-package "$PACKAGE_NAME/pkg/openapi"

popd
