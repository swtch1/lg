package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const appName = "dummy"

func main() {
	r := mux.NewRouter()
	r.PathPrefix("/").Handler(http.HandlerFunc(globalHandler))

	port := os.Getenv("DUMMY_PORT")
	if port == "" {
		port = "8080"
	}
	srv := http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	logrus.Infof("serving on port %s", port)
	logrus.Fatal(srv.ListenAndServe())
}

func globalHandler(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("app", appName).Infof("serving %s", r.URL.Path)

	// introduce a bit of latency
	randomSleep()

	codeHdr := r.Header.Get("CODE")
	code, err := strconv.Atoi(codeHdr)
	if err != nil {
		code = http.StatusBadRequest
	}
	w.WriteHeader(code)

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	fmt.Fprint(w, string(b))
}

func randomSleep() {
	n := rand.Intn(500)
	time.Sleep(time.Millisecond * time.Duration(n))
}
