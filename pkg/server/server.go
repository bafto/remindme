// package server handles the events, notifications
// and listens for incoming calls or updates
package server

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bafto/remindme/pkg/reminder"
	"github.com/gen2brain/beeep"
)

var (
	eventChan    chan reminder.Entry
	missedEvents []reminder.Entry = make([]reminder.Entry, 0)
)

func startEventListeners() error {
	entries, err := reminder.GetAllReminders()
	if err != nil {
		return err
	}

	eventChan = make(chan reminder.Entry, len(entries))

	for _, entry := range entries {
		when, err := entry.GetTime()
		if err != nil {
			return err
		}

		if when.Before(time.Now()) {
			missedEvents = append(missedEvents, entry)
			continue
		}

		go func(event reminder.Entry) {
			<-time.After(when.Sub(time.Now()))
			eventChan <- event
		}(entry)
	}

	return nil
}

func StartServer(port string) (<-chan error, error) {
	if err := startEventListeners(); err != nil {
		return nil, err
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
		case http.MethodDelete:
		default:
			http.Error(w, "invalid request type", http.StatusBadRequest)
		}
	})

	finished := make(chan error, 2)
	go func() {
		for event := range eventChan {
			if err := beeep.Notify(event.Title, event.Msg, ""); err != nil {
				fmt.Fprintf(os.Stderr, "%s", err)
			}
			if _, err := reminder.RemoveReminder(event); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to remove Reminder: %s", err)
			}
		}
	}()
	go func() {
		finished <- http.ListenAndServe(port, nil)
	}()

	return finished, nil
}
