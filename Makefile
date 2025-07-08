run-auth-service:
	cd services/auth-service && air

run-user-service:
	cd services/user-service && air

run:
	$(MAKE) -j2 run-auth-service run-user-service