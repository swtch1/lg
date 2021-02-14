#!/usr/bin/env bash

# this script starts dependencies for running locally

# make sure we're in the correct directory
cd "$(dirname "$0")"

echo 'starting MySQL...'
sudo docker run --rm --name lg_db -d -e MYSQL_ROOT_PASSWORD=$MYSQL_PASS -v $PWD/mysql.conf:/etc/my.cnf -p 127.0.0.1:3306:3306 mysql:8.0.23

echo 'starting redis'
sudo docker run --rm --name lg_cache -d -p 127.0.0.1:6379:6379 redis

echo 'done'
