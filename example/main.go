package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"net/http/httptest"
	_ "net/http/pprof"

	"github.com/felixge/fgprof"
)

const (
	sleepTime   = 10 * time.Millisecond
	cpuTime     = 30 * time.Millisecond
	networkTime = 60 * time.Millisecond
)

// sleepURL is the url for the sleep server used by slowNetworkRequest. It's
// a global variable to keep the cute simplicitly of main's loop.
var sleepURL string

func main() {
	// Run http endpoints for both pprof and fgprof.
	http.DefaultServeMux.Handle("/debug/fgprof", fgprof.Handler())
	go func() {
		addr := "localhost:6060"
		log.Printf("Listening on %s", addr)
		log.Println(http.ListenAndServe(addr, nil))
	}()

	// Start a sleep server to help with simulating slow network requests.
	var stop func()
	sleepURL, stop = StartSleepServer()
	defer stop()

	for i := 0; ; i++ {
		// Http request to a web service that might be slow.
		start := time.Now()
		slowNetworkRequest()
		now := time.Now()
		if i%10000 == 0 {
			fmt.Printf("slowNetworkRequest: %s\n", now.Sub(start))
		}
		// Some heavy CPU computation.
		cpuIntensiveTask()
		now2 := time.Now()
		if i%10000 == 0 {
			fmt.Printf("cpuIntensiveTask: %s\n", now2.Sub(now))
		}
		// Poorly named function that you don't understand yet.
		weirdFunction()
		now3 := time.Now()
		if i%10000 == 0 {
			fmt.Printf("weirdFunction: %s\n", now3.Sub(now2))
		}
	}
}

func slowNetworkRequest() {
	res, err := http.Get(sleepURL + "/?sleep=" + networkTime.String())
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		panic(fmt.Sprintf("bad code: %d", res.StatusCode))
	}
}

func cpuIntensiveTask() {
	start := time.Now()
	for time.Since(start) <= cpuTime {
		// Spend some time in a hot loop to be a little more realistic than
		// spending all time in time.Since().
		for i := 0; i < 1000; i++ {
			_ = i
		}
	}
}

func weirdFunction() {
	time.Sleep(sleepTime)
}

// StartSleepServer starts a server that supports a ?sleep parameter to
// simulate slow http responses. It returns the url of that server and a
// function to stop it.
func StartSleepServer() (url string, stop func()) {
	server := httptest.NewServer(http.HandlerFunc(sleepHandler))
	return server.URL, server.Close
}

func sleepHandler(w http.ResponseWriter, r *http.Request) {
	sleep := r.URL.Query().Get("sleep")
	sleepD, err := time.ParseDuration(sleep)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "bad duration: %s: %s\n", sleep, err)
	}
	time.Sleep(sleepD)
	fmt.Fprintf(w, "slept for: %s\n", sleepD)
}
