#!/usr/bin/env bash

# this script will load the database schema into MySQL.

# make sure we're in the correct directory
cd "$(dirname "$0")"

mysql -h $MYSQL_SERVER --port=$MYSQL_PORT -u$MYSQL_USER -p$MYSQL_PASS -e 'CREATE DATABASE lg;';
mysql -h $MYSQL_SERVER --port=$MYSQL_PORT -u$MYSQL_USER -p$MYSQL_PASS --database lg < schema.sql
