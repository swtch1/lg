#!/usr/bin/env bash

# this file provides local environment varables for testing
# and is meant to be sourced, not executed directly

export MYSQL_SERVER="127.0.0.1"
export MYSQL_USER="root"
export MYSQL_PASS="pass"
export MYSQL_PORT="3306"

export REDIS_HOST="127.0.0.1"
export REDIS_PORT="6379"

export DUMMY_PORT="8080"
export SUT_BASE="127.0.0.1"

export RUN_KEY="RUN01"
export SUT_TARGET_LATENCY_MS="500"
