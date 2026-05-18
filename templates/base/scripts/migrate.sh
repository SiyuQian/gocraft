#!/bin/sh
set -e
migrate -path migrations -database "$DATABASE_URL" "$@"
