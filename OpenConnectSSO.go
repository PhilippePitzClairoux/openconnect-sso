package main

import (
	"flag"
	"github.com/PhilippePitzClairoux/openconnect-sso/internal"
	"log"
)

// flags
var server = flag.String("server", "", "Server to connect to via openconnect")
var username = flag.String("username", "", "Username to inject in login form")
var password = flag.String("password", "", "Password to inject in login form")
var trace = flag.Bool("trace", false, "Enable tracing")

func main() {
	flag.Parse()

	if *server == "" {
		log.Println("missing mandatory parameter --server")
		flag.PrintDefaults()
	}

	openconnect := internal.NewOpenconnectCtx(*server, *username, *password, *trace)

	err := openconnect.Run()
	if err != nil {
		log.Fatal("Could not run openconnect-sso : ", err)
	}
}
