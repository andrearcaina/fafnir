.PHONY: run run-auth-service run-user-service migrate-up migrate-down migrate-status migrate-create

run-auth-service:
	cd services/auth-service && air

run-user-service:
	cd services/user-service && air

run:
	$(MAKE) -j2 run-auth-service run-user-service

migrate-up:
	export $$(cat .env | xargs) && goose -dir shared/pkg/database/migrations up

migrate-down:
	export $$(cat .env | xargs) && goose -dir shared/pkg/database/migrations down

migrate-status:
	export $$(cat .env | xargs) && goose -dir shared/pkg/database/migrations status

migrate-create:
	go tool goose -dir shared/pkg/database/migrations create $(name) sql