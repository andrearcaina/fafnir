.PHONY: default help build run clean

default: help

help:
	@echo "Usage: make <target> [options]"

build:
	@echo "Building logctl..."
	go build -o bin/logctl main.go
	chmod +x bin/logctl
	@echo "Built successfully: bin/logctl"

run:
	@echo "Running logctl..."
	./bin/logctl

clean:
	@echo "Cleaning up..."
	rm -f bin/logctl
	@echo "Cleaned successfully"