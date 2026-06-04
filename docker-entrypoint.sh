#!/bin/bash
set -e

echo "==> Generating /app/config.yaml from template..."
cp /app/config.yaml.template /app/config.yaml

sed -i "s|DB_PASSWORD_PLACEHOLDER|${DB_PASSWORD:-000000}|g" /app/config.yaml
sed -i "s|LLM_API_KEY_PLACEHOLDER|${LLM_API_KEY:-}|g" /app/config.yaml

echo "==> Configuration generated."
exec /app/server
