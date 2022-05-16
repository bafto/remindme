package main

import (
	"fmt"
	"os"

	"github.com/bafto/remindme/pkg/server"
)

func main() {
	if len(os.Args) == 1 {
		if finished, err := server.StartServer(":3050"); err != nil {
			fmt.Println(err)
		} else if err = <-finished; err != nil {
			fmt.Println(err)
		}
	}
}
