package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	quitOnInterrupt()
}

func quitOnInterrupt() {
	// listen for interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)

	// 	sig := <-sigCh
	//
	// 	gracefulWait := time.Second * 10
	// 	d.Log.WithField("signal", sig).Infof("caught interrupt signal, the server will have %v to shutdown gracefully", gracefulWait)
	// 	cancel()
	//
	// 	select {
	// 	case <-time.After(gracefulWait):
	// 		d.Log.Error("graceful timeout expired, exiting anyway")
	// 	case <-doneCh:
	// 		d.Log.Info("graceful server shutdown complete, exiting")
	// 	}
	// 	os.Exit(0)
}
