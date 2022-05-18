package main

import (
	"log"
	"net/http"

	"github.com/bafto/remindme/pkg/client"
	"github.com/bafto/remindme/pkg/server"
)

func serverRunning() bool {
	if resp, err := http.Get("127.0.0:3050"); err != nil {
		return false
	} else {
		defer resp.Body.Close()
		return true
	}
}

func main() {
	if serverRunning() {
		client.StartClient()
	} else {
		if finished, err := server.StartServer(":3050"); err != nil {
			log.Printf("Could not start server: %s\n", err.Error())
		} else if err := <-finished; err != nil {
			log.Printf("Server errored: %s\n", err.Error())
		}
	}
}
