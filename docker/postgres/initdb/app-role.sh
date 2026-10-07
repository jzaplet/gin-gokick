#!/bin/sh
set -eu

for db in "$DB_NAME" "$DB_TEST_NAME"; do
	psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname postgres \
		-v role="$DB_USERNAME" -v password="$DB_PASSWORD" -v db="$db" -f /app-role.sql
done
