#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username postgres <<-EOSQL
	CREATE USER cni WITH PASSWORD 'cni';
	CREATE DATABASE ipam OWNER cni;
	GRANT ALL PRIVILEGES ON DATABASE ipam TO cni;

    \connect ipam
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO cni;
EOSQL

psql -v ON_ERROR_STOP=1 --username cni ipam <<-EOSQL
        CREATE TABLE ips (
        id serial PRIMARY KEY,
        pod_id VARCHAR ( 64 ) UNIQUE NOT NULL,
        interface VARCHAR ( 16 ) UNIQUE NOT NULL,
        ip VARCHAR ( 128 ) UNIQUE NOT NULL,
        created_on TIMESTAMP DEFAULT now()
    );
EOSQL