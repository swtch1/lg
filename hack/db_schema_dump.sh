#!/usr/bin/env bash

# this script will dump the database schema to a file.

# make sure we're in the correct directory
cd "$(dirname "$0")"

rm -f ./schema.sql
mysqldump -h 127.0.0.1 --port=$MYSQL_PORT -u$MYSQL_USER -p$MYSQL_PASS lg > schema.sql
