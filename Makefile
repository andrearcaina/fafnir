.PHONY: default help lint \
		docker-run-auth-service docker-run-user-service docker-run-security-service docker-run-stock-service docker-run-api-gateway \
        docker-prod docker-prod-build docker-stats docker-run docker-build docker-start docker-pause docker-stop docker-status \
        docker-logs docker-nats docker-rm-volumes docker-prune docker-clean docker-reset \
        migrate-up migrate-down migrate-status migrate-create \
        generate seed \
        kube-start kube-stop kube-delete kube-delete-pod kube-deploy kube-reset \
        kube-status kube-nodes kube-pods kube-svc kube-deployments kube-logs \
        kube-forward kube-tunnel \
        locust

default: help

help:
	@echo "Usage: make <target> [options]"

lint:
	@echo "Running linter..."
	@echo "Linting api-gateway..."
	@cd src/api-gateway && golangci-lint run ./...
	@echo "Linting auth-service..."
	@cd src/auth-service && golangci-lint run ./...
	@echo "Linting security-service..."
	@cd src/security-service && golangci-lint run ./...
	@echo "Linting user-service..."
	@cd src/user-service && golangci-lint run ./...
	@echo "Linting stock-service..."
	@cd src/stock-service && golangci-lint run ./...
	@echo "Linting shared..."
	@cd src/shared && golangci-lint run ./...
	@echo "Linting CLI tools..."
	@cd tools/cli/seedctl && golangci-lint run ./...
	@cd tools/cli/logctl && golangci-lint run ./...

# ------------------------------
# Docker Service Operations
# ------------------------------

docker-auth-service:
	./tools/scripts/docker.sh auth-service

docker-user-service:
	./tools/scripts/docker.sh user-service

docker-security-service:
	./tools/scripts/docker.sh security-service

docker-stock-service:
	./tools/scripts/docker.sh stock-service

docker-api-gateway:
	./tools/scripts/docker.sh api-gateway

# ------------------------------
# Docker Lifecycle Operations
# ------------------------------

docker-prod: docker-clean
	./tools/scripts/docker.sh prod

docker-prod-build:
	./tools/scripts/docker.sh build-prod

docker-stats:
	./tools/scripts/docker.sh stats

docker-run:
	./tools/scripts/docker.sh run $(monitoring)

docker-build:
	./tools/scripts/docker.sh build $(monitoring)

docker-start:
	./tools/scripts/docker.sh start

docker-pause:
	./tools/scripts/docker.sh pause

docker-stop:
	./tools/scripts/docker.sh stop

docker-status:
	./tools/scripts/docker.sh status

docker-logs:
	./tools/scripts/docker.sh logs

docker-nats:
	./tools/scripts/docker.sh nats

docker-rm-volumes:
	./tools/scripts/docker.sh rm-volumes

docker-prune:
	./tools/scripts/docker.sh prune

docker-clean: docker-stop docker-prune docker-rm-volumes

docker-reset: docker-clean docker-build docker-run

# ------------------------------
# Database Migrations Operations
# ------------------------------

migrate-up:
	./tools/scripts/migrate.sh up

migrate-down:
	./tools/scripts/migrate.sh down

migrate-status:
	./tools/scripts/migrate.sh status

# make migrate-create db=auth name=seed_some_data
migrate-create:
	./tools/scripts/migrate.sh create $(db) $(name)

# ------------------------------
# Codegen Operations
# ------------------------------

# make generate codegen=<graphql|sqlc>
# if codegen is sqlc, then service is required (service=auth)
generate:
	./tools/scripts/codegen.sh generate $(codegen) $(service)

# ------------------------------
# Database Seed Operations
# ------------------------------

# make seed db=<db_name> (db_name=auth)
# or seed all databases (db=all)
seed:
	cd tools/cli/seedctl && go run main.go --db $(db)

# ------------------------------
# Kubernetes Operations
# ------------------------------

kube-start:
	./tools/scripts/k8s.sh start

kube-stop:
	minikube stop -p fafnir-cluster

kube-delete:
	minikube delete -p fafnir-cluster

kube-uninstall:
	./tools/scripts/k8s.sh uninstall

kube-secrets:
	./tools/scripts/k8s.sh secrets

kube-docker:
	./tools/scripts/k8s.sh docker

kube-deploy:
	./tools/scripts/k8s.sh deploy

kube-upgrade:
	./tools/scripts/k8s.sh upgrade

kube-delete-pod:
	./tools/scripts/k8s.sh delete $(pod)

kube-reset:
	./tools/scripts/k8s.sh reset $(pod)

kube-status:
	./tools/scripts/k8s.sh status

kube-nodes:
	./tools/scripts/k8s.sh nodes

kube-pods:
	./tools/scripts/k8s.sh pods

kube-svc:
	./tools/scripts/k8s.sh svc

kube-deployments:
	./tools/scripts/k8s.sh deployments

kube-logs:
	./tools/scripts/k8s.sh logs $(pod) $(ns)

kube-dashboard:
	minikube dashboard -p fafnir-cluster

kube-ssh:
	minikube ssh -p fafnir-cluster

kube-forward:
	./tools/scripts/k8s.sh forward $(pod)

kube-tunnel:
	minikube tunnel -p fafnir-cluster

# ------------------------------
# Locust Operations
# ------------------------------

locust:
	./tools/scripts/test.sh locust $(users) $(spawn_rate) $(run_time) $(headless)
