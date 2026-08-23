#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE notification_db;
    CREATE DATABASE user_db;
    CREATE DATABASE order_db;
EOSQL