package client

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bafto/remindme/pkg/reminder"
)

type cmdArgs struct {
	title string
	msg   string
	after time.Duration
}

func parseArgs() cmdArgs {
	retArgs := cmdArgs{}

	flag.StringVar(&retArgs.title, "title", "", "title of the reminder")
	flag.StringVar(&retArgs.msg, "msg", "", "message for the reminder")
	days := flag.Int("days", 0, "in how many days the reminder should fire")
	flag.DurationVar(&retArgs.after, "after", 0, "after what duration the reminder should fire")

	flag.Parse()

	retArgs.after += time.Hour * 24 * time.Duration(*days)

	return retArgs
}

func StartClient() {
	args := parseArgs()
	remind := reminder.Entry{
		Title: args.title,
		Msg:   args.msg,
		When:  time.Now().Add(args.after).Format(reminder.TimeLayout),
	}
	if body, err := json.Marshal(remind); err != nil {
		log.Println(err)
		return
	} else {
		buf := bytes.NewBuffer(body)
		resp, err := http.Post("127.0.0.1:3050", "application/json", buf)
		if err != nil {
			log.Println(err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			log.Printf("Server replied with a status of %d: %s\n", resp.StatusCode, string(body))
		}
	}
}
