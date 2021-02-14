package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

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
