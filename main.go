package main

import (
	"log"
	"net/http"

	"github.com/bafto/remindme/pkg/client"
	"github.com/bafto/remindme/pkg/server"
	"github.com/gen2brain/beeep"
)

func serverRunning() bool {
	if resp, err := http.Get("http://127.0.0.1:3050/"); err != nil {
		return false
	} else {
		defer resp.Body.Close()
		return true
	}
}

func main() {
	if serverRunning() {
		if err := client.StartClient(); err != nil {
			log.Println(err)
		}
	} else {
		if finished, err := server.StartServer(":3050"); err != nil {
			log.Printf("Could not start server: %s\n", err.Error())
			beeep.Notify("Remindme Error", err.Error(), "")
		} else if err := <-finished; err != nil {
			log.Printf("Server errored: %s\n", err.Error())
		}
	}
}
