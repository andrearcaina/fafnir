#!/bin/bash

generate() {
  cd "services/api-gateway" && go generate ./...
}

if [ "$1" == "generate" ]; then
  generate
else
  echo "Usage: $0 generate"
  exit 1
fi