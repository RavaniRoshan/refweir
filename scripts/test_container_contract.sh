#!/usr/bin/env bash
set -euo pipefail

# Verifies that Dockerfile and docker-compose.enterprise.yml properly configure dual transports

DOCKERFILE="Dockerfile"
COMPOSE="docker-compose.enterprise.yml"

echo "Checking Dockerfile configuration..."
grep -q "EXPOSE 8080 2222" "$DOCKERFILE" || {
  echo "FAIL: Dockerfile does not EXPOSE 8080 2222"
  exit 1
}

grep -q "REFWEIR_SSH_ADDR=0.0.0.0:2222" "$DOCKERFILE" || {
  echo "FAIL: Dockerfile missing REFWEIR_SSH_ADDR=0.0.0.0:2222"
  exit 1
}

echo "Checking docker-compose.enterprise.yml configuration..."
grep -q "2222:2222" "$COMPOSE" || {
  echo "FAIL: docker-compose.enterprise.yml missing port 2222 mapping"
  exit 1
}

grep -q "v0.6.0-beta" "$COMPOSE" || {
  echo "FAIL: docker-compose.enterprise.yml missing current version tag v0.6.0-beta"
  exit 1
}

echo "PASS: Container contracts verified"
