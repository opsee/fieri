#!/bin/bash
set -e

echo "loading schema for tests..."
echo "drop database if exists fieri_test; create database fieri_test" | psql -U postgres -h postgresql
migrate -url "$POSTGRES_CONN" -path ./migrations up
