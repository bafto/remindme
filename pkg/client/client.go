package client

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bafto/remindme/pkg/reminder"
)

type cmdArgs struct {
	title  string
	msg    string
	after  time.Duration
	list   bool
	remove int
}

func printCmdUsage() {
	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage()
}

func parseArgs() cmdArgs {
	retArgs := cmdArgs{}

	flag.StringVar(&retArgs.title, "title", "", "title of the reminder")
	flag.StringVar(&retArgs.msg, "msg", "", "message for the reminder")
	days := flag.Int("days", 0, "in how many days the reminder should fire")
	flag.DurationVar(&retArgs.after, "after", 0, "after what duration the reminder should fire (Format XhYmZs)")
	flag.BoolVar(&retArgs.list, "list", false, "display a list of pending reminders")
	flag.IntVar(&retArgs.remove, "remove", -1, "remove the reminder at the given index (see $remindme -list for indices)")

	flag.Parse()

	retArgs.after += time.Hour * 24 * time.Duration(*days)

	return retArgs
}

func printReminders() error {
	if entries, err := reminder.GetAllRemindersSorted(); err != nil {
		return err
	} else {
		fmt.Printf("%d reminders:\n\n", len(entries))
		for i := range entries {
			when, _ := entries[i].GetTime()
			fmt.Printf("(%d) %s %s:\n\t%s\n", i, when.Format("2006-01-02 15:04:05"), entries[i].Title, entries[i].Msg)
		}
	}
	return nil
}

func StartClient() error {
	args := parseArgs()
	if len(os.Args) == 1 {
		printCmdUsage()
		return nil
	}
	if args.list {
		if err := printReminders(); err != nil {
			log.Printf("Error while printing reminders: %s\n", err.Error())
		}
		return nil
	}
	if args.remove >= 0 {
		entries, err := reminder.GetAllRemindersSorted()
		if err != nil {
			return err
		}
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://127.0.0.1:3050/?id=%s", entries[args.remove].Id.String()), nil)
		if err != nil {
			return err
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			log.Printf("Server replied with a status code of %d: %s\n", resp.StatusCode, string(body))
			return nil
		}
		return nil
	}
	remind := reminder.NewEntry(
		time.Now().Add(args.after),
		args.title,
		args.msg,
	)
	if body, err := json.MarshalIndent(remind, "", "\t"); err != nil {
		return err
	} else {
		buf := bytes.NewReader(body)
		resp, err := http.Post("http://127.0.0.1:3050/", "application/json", buf)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			log.Printf("Server replied with a status of %d: %s\n", resp.StatusCode, string(body))
		}
	}
	return nil
}
