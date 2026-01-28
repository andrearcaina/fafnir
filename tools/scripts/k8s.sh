#!/bin/bash

set -e

ENV_FILE="infra/env/.env.dev"
LOCAL_IMAGES=(fafnir-auth-service fafnir-user-service fafnir-security-service fafnir-stock-service fafnir-api-gateway)

case "$1" in
  start)
    # start a minikube cluster called fafnir-cluster with 1 control plane and 2 worker nodes
    # and utilizing the docker driver, with 2GB memory (RAM) and 2 CPUs
    # so a total of 3 nodes, 6GB RAM and 6 CPUs allocated for the cluster
    if ! minikube status -p fafnir-cluster &>/dev/null; then
      echo "Starting minikube cluster 'fafnir-cluster'..."
      minikube start -p fafnir-cluster --driver=docker --memory=2048 --cpus=2 --nodes=3
      echo "Waiting for nodes to be ready..."
      sleep 10 # wait for nodes to be ready
    else
      echo "Minikube cluster 'fafnir-cluster' is already running."
    fi

    # switch to the fafnir-cluster context
    minikube profile fafnir-cluster
    kubectl config use-context fafnir-cluster
    
    echo "Applying node labels..."
    kubectl label node fafnir-cluster logging-exclude=false --overwrite || true
    kubectl label node fafnir-cluster-m02 logging-exclude=false --overwrite || true
    kubectl label node fafnir-cluster-m03 logging-exclude=false --overwrite || true

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

    # load local docker images into minikube
    for image in "${LOCAL_IMAGES[@]}"; do
      echo "Loading image $image into minikube..."
      minikube image load "$image:latest" -p fafnir-cluster
    done

    echo "Minikube multi-node Kubernetes cluster setup completed."
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
      minikube image load "$image:latest" -p fafnir-cluster
    done
    ;;
  deploy)
    # installs the helm chart if not exists (for first time)
    echo "Deploying fafnir using helm chart..."
    helm upgrade --install dev deployments/helm/fafnir --namespace fafnir --create-namespace
    ;;
  upgrade)
    # upgrades the helm chart if it exists
    echo "Upgrading fafnir using helm chart..."
    helm upgrade dev deployments/helm/fafnir --namespace fafnir
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
        kubectl port-forward -n fafnir svc/api-gateway 8080:8080
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
