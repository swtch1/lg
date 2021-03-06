package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-redis/redis/v7"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	redis_feed "github.com/swtch1/lg/feed/redis"
	"github.com/swtch1/lg/loadgen"
	"github.com/swtch1/lg/store/mysql"
	"github.com/swtch1/lg/store/redisdb"
)

const appName = "stressor"

func main() {
	// get config details
	var cfg config
	if err := cfg.setFromEnv(); err != nil {
		logrus.Fatal(err)
	}

	log := mustNewLog(cfg)
	log = log.WithField("app", appName)

	var gen *loadgen.Generator
	{
		// set a maximum wait time that makes sense in case the service cannot find a scale factor
		tgtLatency, err := strconv.Atoi(cfg.sutTargetLatency)
		if err != nil {
			log.Fatal("target latency must be an integer, the target latency in MS for the SUT")
		}
		maxWait := (time.Millisecond * time.Duration(tgtLatency) * 3)

		rc := mustNewRedisClient(cfg, log)
		feed := redis_feed.NewFeed(rc)
		lw := mustNewLatencyDB(cfg, log)
		sdb := redisdb.NewScaleDB(rc, cfg.runKey)
		gen = loadgen.NewGenerator(cfg.sut_base, maxWait, feed, lw, sdb, log)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// run the generator for a time
	go func() {
		gen.Run(ctx, 3)
		log.Fatal("generator run terminated")
	}()

	// listen for interrupts
	quitOnInterrupt(cancel, log)
}

func mustNewRedisClient(cfg config, log *logrus.Entry) *redis.Client {
	rc, err := redisdb.NewClient(
		redisdb.Config{
			Host: cfg.redisHost,
			Port: cfg.redisPort,
		},
		0,
	)
	if err != nil {
		log.WithError(err).Fatal("failed to create redis client")
	}

	return rc
}

func mustNewLatencyDB(cfg config, log *logrus.Entry) *mysql.LatencyDB {
	readDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true",
		cfg.dbUser,
		cfg.dbPass,
		cfg.dbServer,
		cfg.dbPort,
	)
	db, err := sqlx.Open("mysql", readDSN)
	if err != nil {
		log.WithError(err).Fatalf("unable to open database")
	}
	if err := db.Ping(); err != nil {
		log.WithError(err).Fatalf("unable to ping database")
	}
	log.Infof("successfully connected to database")

	db.SetMaxIdleConns(0)
	db.SetMaxOpenConns(15)
	db.SetConnMaxLifetime(time.Minute * time.Duration(5))

	return mysql.NewLatencyDB(db)
}

func mustNewLog(cfg config) *logrus.Entry {
	log := logrus.New()
	level, err := logrus.ParseLevel(cfg.logLevel)
	if err != nil {
		logrus.Fatal(err)
	}
	log.Level = level
	rc := strings.ToLower(cfg.logReportCaller)
	if rc == "y" || rc == "1" || rc == "true" {
		log.ReportCaller = true
	}
	return logrus.NewEntry(log)
}

func quitOnInterrupt(cancel context.CancelFunc, log *logrus.Entry) {
	// listen for interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)

	sig := <-sigCh

	gracefulWait := time.Second * 10
	log.WithField("signal", sig).Infof("caught interrupt signal, the server will have %v to shutdown gracefully", gracefulWait)
	cancel()

	<-time.After(gracefulWait)
}
