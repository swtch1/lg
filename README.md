# lg

Distributed load generation.

## Quick Start

Commands are assumed to be running from a Linux BASH shell or similar.

- source the local environment variables with `source ./hack/local_env_vars-source.sh`
- start dependencies with `./hack/run_mysql.sh`
- wait until the mysql database is live. basic verification can be done with `watch 'mysql -h 127.0.0.1 -u$MYSQL_USER -p$MYSQL_PASS -e "exit" && echo "ready" || echo "not ready"'`
- create mysql schema with `./hack/db_schema_load.sh`
- run [dummy server](./cmd/dummy) with `go run ./cmd/dummy &`
- run [coordinator](./cmd/coordinator) with `go run ./cmd/coordinator &`
- run [stressor](./cmd/stressor) with `go run ./cmd/stressor &`

## Components

- The stressor reads request/response pairs from a data store and generates load with those pairs directed at the SUT (system under test).
- The coordinator reads aggregated metrics from all stressors and adjusts the scale factor, creating a feedback loop for the stressors.
- The dummy server represents a SUT and just responds with whatever it's given.

In this particular test implementation the stressors write their aggregated metrics to a MySQL database. They would gather rrpairs from Redis but to save time they are just pulling pairs from a local slice. The coordinator reads the last n metrics from the database and uses it to determine the scale factor.  For simplicity the scale factor is just updated in, and read from, a single redis value.  This provides a scalable delivery mechanism with little extra code.
