	package main

	import (
	    "fmt"
	    "log"
	)

	func run() error {
	    fmt.Println("Starting noti-rabbitmq service...")
	    return nil
	}

	func main() {
	    if err := run(); err != nil {
	        log.Fatal(err)
	    }
	    fmt.Println("Service initialized successfully")
	}

