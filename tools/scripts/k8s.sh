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
    else
      echo "Minikube cluster 'fafnir-cluster' is already running."
    fi

    # switch to the fafnir-cluster context
    minikube profile fafnir-cluster
    kubectl config use-context fafnir-cluster
    kubectl label node fafnir-cluster logging-exclude=false --overwrite
    kubectl label node fafnir-cluster-m02 logging-exclude=false --overwrite
    kubectl label node fafnir-cluster-m03 logging-exclude=false --overwrite

    # create fafnir namespace (for app)
    kubectl create namespace fafnir --dry-run=client -o yaml | kubectl apply -f -

    # create logging namespace (for fafnir logs)
    kubectl create namespace logging --dry-run=client -o yaml | kubectl apply -f -

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
  deploy)
    if [[ "$2" == "all" ]]; then
      echo "Deploying all resources..."
      kubectl apply -f deployments/k8s/ --recursive
    else
      echo "Deploying $2..."
      kubectl apply -f deployments/k8s/deployment/$2.yaml
    fi
    ;;
  delete)
    kubectl delete -f deployments/k8s/ --recursive
    ;;
  reset)
    if [[ "$2" == "all" ]]; then
      echo "Restarting all deployments..."
      kubectl rollout restart deployment -n fafnir  
      kubectl rollout restart deployment -n logging
    else
      echo "Restarting $2..."
      kubectl rollout restart deployment/$2 -n fafnir
    fi
    ;;
  status)
    for ns in fafnir logging; do
      echo "=== Namespace: $ns ==="
      kubectl get all -n $ns -o wide
    done
    ;;
  nodes)
    kubectl get nodes -o wide
    ;;
  pods)
    for ns in fafnir logging; do
      echo "=== Namespace: $ns ==="
      kubectl get pods -n $ns -o wide
    done
    ;;
  svc)
    for ns in fafnir logging; do
      echo "=== Namespace: $ns ==="
      kubectl get svc -n $ns -o wide
    done
    ;;
  deployments)
    for ns in fafnir logging; do
      echo "=== Namespace: $ns ==="
      kubectl get deployments -n $ns -o wide
    done
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
        kubectl port-forward -n fafnir svc/fafnir-api-gateway 8080:80
        ;;
      ps)
        kubectl port-forward -n fafnir svc/fafnir-postgres 5432:5432
        ;;
      es)
        kubectl port-forward -n logging svc/elasticsearch 9200:9200
        ;;
      *)
        echo "Only postgres (ps), api-gateway (ag), and elasticsearch (es) supported for port forwarding."
        exit 1
        ;;
    esac
    ;;
  *)
    echo "Usage: $0 {start|deploy|delete|reset|status|nodes|pods|svc|deployments|logs|forward}"
esac