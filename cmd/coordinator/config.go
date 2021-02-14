package main

import (
	"fmt"
	"os"
)

// envConfig is configuration from the environment.
type config struct {
	logLevel        string
	logReportCaller string

	dbServer string
	dbUser   string
	dbPass   string
	dbPort   string

	redisHost string
	redisPort string

	// runKey helps keep multiple runs from interfering with each other. This should be unique for each run.
	runKey           string
	sutTargetLatency string
}

type assignment struct {
	val        *string
	envVar     string
	defaultVal interface{}
}

// setFromEnv sets values from environment variables, with optional defaults.
func (c *config) setFromEnv() error {
	var errs []error

	errs = append(errs, set(assignment{&c.logLevel, "LOG_LEVEL", "info"}))
	errs = append(errs, set(assignment{&c.logReportCaller, "LOG_REPORT_CALLER", "true"}))
	errs = append(errs, set(assignment{&c.dbServer, "MYSQL_SERVER", nil}))
	errs = append(errs, set(assignment{&c.dbUser, "MYSQL_USER", nil}))
	errs = append(errs, set(assignment{&c.dbPass, "MYSQL_PASS", nil}))
	errs = append(errs, set(assignment{&c.dbPort, "MYSQL_PORT", nil}))
	errs = append(errs, set(assignment{&c.redisHost, "REDIS_HOST", nil}))
	errs = append(errs, set(assignment{&c.redisPort, "REDIS_PORT", "6379"}))
	errs = append(errs, set(assignment{&c.runKey, "RUN_KEY", "6379"}))
	errs = append(errs, set(assignment{&c.runKey, "SUT_TGT_LATENCY_MS", nil}))

	// this format allows us to report on every missing
	// configuration element rather than just the next one
	var es string
	for _, err := range errs {
		if err != nil {
			es += err.Error() + "\n"
		}
	}
	if es != "" {
		return fmt.Errorf(es)
	}
	return nil
}

// set the associated value and return an error if a value was expected but not
// given, and no default exists.
func set(a assignment) error {
	var ok bool
	*a.val, ok = os.LookupEnv(a.envVar)
	if !ok {
		if a.defaultVal == nil {
			return fmt.Errorf("value for env var %s was required but not given", a.envVar)
		}
		*a.val = a.defaultVal.(string)
		return nil
	}
	return nil
}
