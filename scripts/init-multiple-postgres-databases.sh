#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE ecommerce_db;
    CREATE DATABASE notification_db;
    CREATE DATABASE user_db;
EOSQL