package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	flag "github.com/spf13/pflag"
)

var (
	noDup    bool
	authConf string
	port     string
	conf     *Config
)

func init() {
	flag.BoolVar(&noDup, "n", false, "Stops duplicate items from being queued")
	flag.StringVar(&authConf, "auth", "", "Enables Active Directory™ authentication")
	flag.StringVar(&port, "port", "8080", "Port to serve the webserver on")
	flag.Parse()
}

func main() {
	var (
		ctl chan string   = nil
		q   chan []string = nil
	)
	useAuth := false
	if authConf != "" {
		useAuth = true
	}
	ch := make(chan string, 20)

	if useAuth {
		ctl = make(chan string, 2)
		q = make(chan []string, 1)
		http.HandleFunc("/signin", handleSignin)
		if cf, err := os.Open(authConf); err != nil {
			fmt.Fprintf(os.Stderr, "Could not open conf file: %v\n", err)
			os.Exit(1)
		} else {
			if conf, err = readConf(cf); err != nil {
				fmt.Fprintf(os.Stderr, "Could not parse conf file: %v\n", err)
				os.Exit(1)
			}
		}
	}

	http.HandleFunc("/", genIndexHandle(ch, ctl, q, useAuth))
	go DJ(ch, ctl, q, noDup)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
