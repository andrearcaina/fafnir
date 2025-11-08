#!/bin/bash

set -e

ENV_FILE="infra/env/.env.dev"
LOCAL_IMAGES=(fafnir-auth-service fafnir-user-service fafnir-security-service fafnir-stock-service fafnir-api-gateway)

case "$1" in
  setup)
    # create namespace
    kubectl create namespace fafnir --dry-run=client -o yaml | kubectl apply -f -

    # create or update secrets
    if kubectl get secret fafnir-secrets -n fafnir &>/dev/null; then
      echo "Secret fafnir-secrets exists, updating..."
      kubectl delete secret fafnir-secrets -n fafnir
    fi

    kubectl create secret generic fafnir-secrets \
      --from-env-file=$ENV_FILE \
      --namespace=fafnir

    # verify creation
    kubectl get secret fafnir-secrets -n fafnir
    kubectl describe secret fafnir-secrets -n fafnir

    # load local docker images into minikube
    for image in "${LOCAL_IMAGES[@]}"; do
      echo "Loading image $image into minikube..."
      minikube image load "$image:latest"
    done

    echo "Minikube setup completed."
    ;;
  deploy)
    if [[ "$2" == "all" ]]; then
      echo "Deploying all resources..."
      kubectl apply -f deployments/k8s/ -n fafnir --recursive
    else
      echo "Deploying $2..."
      kubectl apply -f deployments/k8s/deployment/$2.yaml -n fafnir
    fi
    ;;
  delete)
    kubectl delete -f deployments/k8s/ -n fafnir --recursive
    ;;
  reset)
    if [[ "$2" == "all" ]]; then
      echo "Restarting all deployments..."
      kubectl rollout restart deployment -n fafnir
    else
      echo "Restarting $2..."
      kubectl rollout restart deployment/$2 -n fafnir
    fi
    ;;
  status)
    kubectl get all -n fafnir
    ;;
  pods)
    kubectl get pods -n fafnir
    ;;
  svc)
    kubectl get svc -n fafnir
    ;;
  deployments)
    kubectl get deployments -n fafnir
    ;;
  logs)
      [[ -z "$2" ]] && { echo "App name required for logs."; exit 1; }
      pod=$(kubectl get pods -n fafnir -l app="$2" -o jsonpath='{.items[0].metadata.name}')
      kubectl logs -n fafnir "$pod" --follow
      ;;

  forward)
    if [[ "$2" != "postgres" ]]; then
      echo "postgres is the only supported service for port-forwarding."
      exit 1;
    fi
    kubectl -n fafnir port-forward svc/postgres 5432:5432
    ;;
  *)
    echo "Usage: $0 {setup|deploy|delete|reset|status|pods|svc|deployments|logs|forward}"
esac