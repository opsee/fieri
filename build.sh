#!/bin/bash
set -e

echo "loading schema for tests..."
echo "drop database if exists fieri_test; create database fieri_test" | psql $POSTGRES_CONN
migrate -url "$POSTGRES_CONN" -path ./migrations up
