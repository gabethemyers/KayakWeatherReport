#!/usr/bin/env bash
set -e

# Ensure we're in the project directory
PROJECT_DIR="$HOME/KayakWeatherReport"
if [ -d "$PROJECT_DIR" ]; then
    cd "$PROJECT_DIR"
fi

echo "==> Pulling latest changes from git..."
git pull

echo "==> Stopping existing production stack..."
docker compose --profile production down

echo "==> Rebuilding and starting production stack..."
docker compose --profile production up -d --build

echo "==> Pruning dangling build images..."
docker image prune -f

echo "==> Deployment finished successfully!"
docker compose --profile production ps
