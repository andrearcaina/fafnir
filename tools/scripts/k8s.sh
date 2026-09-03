#!/bin/bash

set -e

ENV_FILE="infra/env/.env.dev"
LOCAL_IMAGES=(fafnir-auth-service fafnir-user-service fafnir-security-service fafnir-stock-service fafnir-api-gateway fafnir-order-service fafnir-portfolio-service)
CLUSTER_NAME="fafnir-cluster"
KIND_CONFIG="deployments/kind/fafnir-cluster.yaml"

case "$1" in
  start)
    if ! kind get clusters | grep -qx "$CLUSTER_NAME"; then
      echo "Starting kind cluster '$CLUSTER_NAME'..."
      kind create cluster --name "$CLUSTER_NAME" --config "$KIND_CONFIG"
    else
      echo "Kind cluster '$CLUSTER_NAME' already exists."
    fi

    kubectl config use-context "kind-$CLUSTER_NAME"
    kubectl wait --for=condition=Ready nodes --all --timeout=120s

    echo "Applying node labels..."
    kubectl get nodes -o name | xargs -r -I{} kubectl label {} logging-exclude=false --overwrite

    # create fafnir namespace (for app and infra)
    echo "Creating 'fafnir' namespace..."
    kubectl create namespace fafnir --dry-run=client -o yaml | kubectl apply -f -

    # create or update secrets
    if kubectl get secret fafnir-secrets -n fafnir &>/dev/null; then
      echo "Secret fafnir-secrets exists, updating..."
      kubectl delete secret fafnir-secrets -n fafnir
    fi

    kubectl create secret generic fafnir-secrets --from-env-file=$ENV_FILE --namespace=fafnir

    # verify creation
    kubectl get secret fafnir-secrets -n fafnir
    kubectl describe secret fafnir-secrets -n fafnir

    # load local docker images into kind
    for image in "${LOCAL_IMAGES[@]}"; do
      echo "Loading image $image into kind..."
      kind load docker-image "$image:latest" --name "$CLUSTER_NAME"
    done

    echo "Kind multi-node Kubernetes cluster setup completed."
    ;;
  secrets)
    echo "Updating secrets..."
    kubectl delete secret fafnir-secrets -n fafnir
    kubectl create secret generic fafnir-secrets --from-env-file=$ENV_FILE --namespace=fafnir
    ;;
  docker)
    echo "Updating local docker images..."
    for image in "${LOCAL_IMAGES[@]}"; do
      echo "Updating image $image..."
      kind load docker-image "$image:latest" --name "$CLUSTER_NAME"
    done
    ;;
  deploy)
    # use vendored dependency archives when available
    # Build them only for a fresh checkout or when a dependency archive is missing
    dependencies_ready=true
    for dependency in postgresql redis nats loki promtail traefik; do
      if ! compgen -G "deployments/helm/fafnir/charts/${dependency}-*.tgz" >/dev/null; then
        dependencies_ready=false
        break
      fi
    done

    if [ "$dependencies_ready" = false ]; then
      helm repo add bitnami https://charts.bitnami.com/bitnami >/dev/null 2>&1 || true
      helm repo add nats https://nats-io.github.io/k8s/helm/charts/ >/dev/null 2>&1 || true
      helm repo add grafana https://grafana.github.io/helm-charts >/dev/null 2>&1 || true
      helm repo add traefik https://traefik.github.io/charts >/dev/null 2>&1 || true
      echo "Building missing Helm chart dependencies..."
      helm dependency build deployments/helm/fafnir
    fi

    echo "Deploying fafnir using Helm..."
    helm upgrade --install dev deployments/helm/fafnir --namespace fafnir --create-namespace
    ;;
  upgrade)
    # upgrades the helm chart if it exists
    echo "Upgrading fafnir using helm chart..."
    if [ "${2:-true}" = "false" ]; then
      helm upgrade dev deployments/helm/fafnir --namespace fafnir --no-hooks
    else
      helm upgrade dev deployments/helm/fafnir --namespace fafnir
    fi
    ;;
  delete)
    [[ -z "$2" ]] && { echo "App name required for delete."; exit 1; }
    if [ "$2" == "all" ]; then
      echo "Deleting all pods in fafnir namespace..."
      kubectl delete pods --all -n fafnir
    fi
    namespace="${3:-fafnir}" # defaults to fafnir if not provided
    pod=$(kubectl get pods -n "$namespace" -l app="$2" -o jsonpath='{.items[0].metadata.name}')
    kubectl delete pod "$pod" -n "$namespace"
    ;;
  uninstall)
    echo "Uninstalling fafnir using helm chart..."
    helm uninstall dev -n fafnir
    ;;
  reset)
    echo "Restarting all deployments in fafnir namespace..."
    kubectl rollout restart deployment -n fafnir
    kubectl rollout restart statefulset -n fafnir
    kubectl rollout restart daemonset -n fafnir
    ;;
  status)
    echo "=== Namespace: fafnir ==="
    kubectl get all -n fafnir -o wide
    ;;
  nodes)
    kubectl get nodes -o wide
    ;;
  pods)
    echo "=== Namespace: fafnir ==="
    kubectl get pods -n fafnir -o wide
    ;;
  svc)
    echo "=== Namespace: fafnir ==="
    kubectl get svc -n fafnir -o wide
    ;;
  deployments)
    echo "=== Namespace: fafnir ==="
    kubectl get deployments -n fafnir -o wide
    ;;
  logs)
    [[ -z "$2" ]] && { echo "App name required for logs."; exit 1; }
    namespace="${3:-fafnir}" # defaults to fafnir if not provided
    pod=$(kubectl get pods -n "$namespace" -l app="$2" -o jsonpath='{.items[0].metadata.name}')
    kubectl logs -n "$namespace" "$pod" --follow
    ;;
  forward)
    case "$2" in
      ag)
        echo "API Gateway is available at http://localhost (Traefik ingress)."
        ;;
      ps)
        kubectl port-forward -n fafnir svc/postgres 5432:5432
        ;;
      loki)
        kubectl port-forward -n fafnir svc/loki 3100:3100
        ;;
      *)
        echo "Supported services: ag (api-gateway), ps (postgres), loki"
        exit 1
        ;;
    esac
    ;;
  *)
    echo "Usage: $0 {start|secrets|docker|deploy|upgrade|delete|uninstall|reset|status|nodes|pods|svc|deployments|logs|forward}"
esac
