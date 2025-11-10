.PHONY: default help \
		run-auth-service run-user-service run-security-service run-stock-service run-api-gateway \
        run build start pause stop status logs rm-volumes prune clean reset \
        migrate-up migrate-down migrate-status migrate-create \
        generate seed \
        kube-start kube-stop kube-deploy kube-delete kube-reset \
        kube-status kube-nodes kube-pods kube-svc kube-deployments kube-logs \
        kube-forward kube-tunnel

default: help

help:
	@echo "Usage: make <target> [options]"

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

docker-run:
	./tools/scripts/docker.sh run $(monitoring)

docker-build:
	./tools/scripts/docker.sh build

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
	cd tools/seeder && go run main.go --db $(db)

# ------------------------------
# Kubernetes Operations
# ------------------------------

kube-start:
	./tools/scripts/k8s.sh start

kube-stop:
	minikube stop -p fafnir-cluster

kube-deploy:
	./tools/scripts/k8s.sh deploy $(pod)

kube-delete:
	./tools/scripts/k8s.sh delete

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
	./tools/scripts/k8s.sh logs $(pod)

kube-dashboard:
	minikube dashboard -p fafnir-cluster

kube-ssh:
	minikube ssh -p fafnir-cluster

kube-forward:
	./tools/scripts/k8s.sh forward $(pod)

kube-tunnel:
	minikube tunnel -p fafnir-cluster