#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

KUTTL_VERSION="0.15.0"
arch_name="$(uname -m)"

if [[ $arch_name == "arm64" ]]; then
    curl -L https://github.com/kudobuilder/kuttl/releases/download/v${KUTTL_VERSION}/kubectl-kuttl_${KUTTL_VERSION}_darwin_arm64 --output kuttl
elif [[ $(uname -s) == "Darwin" ]]; then
    curl -L https://github.com/kudobuilder/kuttl/releases/download/v${KUTTL_VERSION}/kubectl-kuttl_${KUTTL_VERSION}_darwin_x86_64 --output kuttl
elif [[ "$(expr substr $(uname -s) 1 5)" == "Linux" ]]; then
    curl -L https://github.com/kudobuilder/kuttl/releases/download/v${KUTTL_VERSION}/kubectl-kuttl_${KUTTL_VERSION}_linux_x86_64 --output kuttl
fi

chmod +x kuttl

./kuttl test --config test/e2e/e2e-test.yaml
