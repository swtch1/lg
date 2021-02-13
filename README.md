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

## Scaling Algorithm

Right now the scaling algorithm is simple but the system supports tuning over time with little modification.  Again for simplicity, the scale factor used here is just a sleep time in ms.  If the scale factor is 100 then each request will sleep for 100 ms before being processed.

We start with a number of stressors running as separate goroutines and a single coordinator. Each stressor makes requests, gathers data, and updates a MySQL database, taking the scale factor into account.  The coordinator reads the latest records from MySQL, uses the data to determine a scale factor and publishes that factor where stressors can read it.

Assume we have a SUT with an initial scale factor `f`.  We define a maximum acceptable latency `m` for the SUT.  We define an upper bound `ub` and lower bound `lb` for `f`, which start at `3m` and `0` respectively.  Starting at `3m` throughput should start low since `f` is translated directly into wait time between requests.

The coordinator assess current, averaged SUT latency `c` against `m` and adjusts `ub` and `lb` accordingly, as described in the pseudo code below.

```go
const m = getM()
ub, lb := 3*m, 0
c := getC()
if c < m {
    f--
    ub = f
} else {
    f++
    lb = f
}
```

Adjusting `ub` and `lb` help drive the SUT toward maximum throughput with acceptable response time, which not allowing `f` to ping pong wildly past bounds that have already been tested.
