#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

KUTTL_VERSION="0.7.2"

if [[ "$(uname)" == "Darwin" ]]; then
    curl -L https://github.com/kudobuilder/kuttl/releases/download/v${KUTTL_VERSION}/kubectl-kuttl_${KUTTL_VERSION}_darwin_x86_64 --output kuttl
elif [[ "$(expr substr $(uname -s) 1 5)" == "Linux" ]]; then
    curl -L https://github.com/kudobuilder/kuttl/releases/download/v${KUTTL_VERSION}/kubectl-kuttl_${KUTTL_VERSION}_linux_x86_64 --output kuttl
fi

chmod +x kuttl

./kuttl test --config test/e2e/e2e-test.yaml