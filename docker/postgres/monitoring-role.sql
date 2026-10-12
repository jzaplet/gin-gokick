SELECT format('CREATE ROLE %I LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE PASSWORD %L', :'role', :'password')
WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = :'role') \gexec

GRANT pg_monitor TO :"role";

CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
