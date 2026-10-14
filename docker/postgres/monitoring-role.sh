#!/bin/sh
set -eu

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$DB_NAME" \
	-v role=monitoring -v password="$(cat "$DB_MONITORING_PASSWORD_FILE")" -f /monitoring-role.sql
