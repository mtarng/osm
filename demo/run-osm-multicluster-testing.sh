#!/bin/bash

set -aueo pipefail

if [ ! -f .env ]; then
    echo -e "\nThere is no .env file in the root of this repository."
    echo -e "Copy the values from .env.example into .env."
    echo -e "Modify the values in .env to match your setup.\n"
    echo -e "    cat .env.example > .env\n\n"
    exit 1
fi

# shellcheck disable=SC1091
source .env

# Set meaningful defaults for env vars we expect from .env
CI="${CI:-false}"  # This is set to true by Github Actions
MESH_NAME="${MESH_NAME:-osm}"
K8S_NAMESPACE="${K8S_NAMESPACE:-osm-system}"
BOOKBUYER_NAMESPACE="${BOOKBUYER_NAMESPACE:-bookbuyer}"
BOOKSTORE_NAMESPACE="${BOOKSTORE_NAMESPACE:-bookstore}"
BOOKTHIEF_NAMESPACE="${BOOKTHIEF_NAMESPACE:-bookthief}"
BOOKWAREHOUSE_NAMESPACE="${BOOKWAREHOUSE_NAMESPACE:-bookwarehouse}"
CERT_MANAGER="${CERT_MANAGER:-tresor}"
CTR_REGISTRY="${CTR_REGISTRY:-localhost:5000}"
CTR_REGISTRY_CREDS_NAME="${CTR_REGISTRY_CREDS_NAME:-acr-creds}"
DEPLOY_TRAFFIC_SPLIT="${DEPLOY_TRAFFIC_SPLIT:-true}"
CTR_TAG="${CTR_TAG:-$(git rev-parse HEAD)}"
IMAGE_PULL_POLICY="${IMAGE_PULL_POLICY:-Always}"
ENABLE_DEBUG_SERVER="${ENABLE_DEBUG_SERVER:-true}"
ENABLE_EGRESS="${ENABLE_EGRESS:-false}"
DEPLOY_GRAFANA="${DEPLOY_GRAFANA:-false}"
DEPLOY_JAEGER="${DEPLOY_JAEGER:-false}"
ENABLE_FLUENTBIT="${ENABLE_FLUENTBIT:-false}"
DEPLOY_PROMETHEUS="${DEPLOY_PROMETHEUS:-false}"
DEPLOY_WITH_SAME_SA="${DEPLOY_WITH_SAME_SA:-false}"
ENVOY_LOG_LEVEL="${ENVOY_LOG_LEVEL:-debug}"
DEPLOY_ON_OPENSHIFT="${DEPLOY_ON_OPENSHIFT:-false}"

# For any additional installation arguments. Used heavily in CI.
optionalInstallArgs=$*

exit_error() {
    error="$1"
    echo "$error"
    exit 1
}

# Check if Docker daemon is running
docker info > /dev/null || { echo "Docker daemon is not running"; exit 1; }

make build-osm

# TODO - orignal cluster cleanup? Optionally check if they exist, then clean up.
# cleanup stale resources from previous runs
./demo/clean-kubernetes.sh









# Create Cluster 1
CLUSTER_ONE_NAME="multicluster-osm-1"
KIND_CLUSTER_NAME=$CLUSTER_ONE_NAME
echo "Creating first cluster named: $KIND_CLUSTER_NAME"
./scripts/kind-with-registry.sh

# Create Cluster 1
CLUSTER_TWO_NAME="multicluster-osm-2"
KIND_CLUSTER_NAME=$CLUSTER_TWO_NAME
echo "Creating second cluster named: $KIND_CLUSTER_NAME"
./scripts/kind-with-registry.sh


# Install on cluster 1
kubectl cluster-info --context kind-$CLUSTER_ONE_NAME

# Install on cluster 2
kubectl cluster-info --context kind-$CLUSTER_TWO_NAME
