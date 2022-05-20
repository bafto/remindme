// package server handles the events, notifications
// and listens for incoming calls or updates
package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bafto/remindme/pkg/reminder"
	"github.com/gen2brain/beeep"
	"github.com/google/uuid"
)

var (
	eventChan chan reminder.Entry
	// to close running event goroutines on delete
	tickerChannels map[uuid.UUID]chan bool = make(map[uuid.UUID]chan bool)
	missedEvents   []reminder.Entry        = make([]reminder.Entry, 0)
)

func queueEvent(event reminder.Entry) error {
	when, err := event.GetTime()
	if err != nil {
		return err
	}

	if when.Before(time.Now()) {
		missedEvents = append(missedEvents, event)
		return nil
	}

	tickerChannels[event.Id] = make(chan bool)
	go func(event reminder.Entry) {
		select {
		case <-time.After(time.Until(when)):
			eventChan <- event
		case <-tickerChannels[event.Id]:
			return
		}
	}(event)

	return nil
}

func startEventListeners() error {
	entries, err := reminder.GetAllReminders()
	if err != nil {
		return err
	}

	eventChan = make(chan reminder.Entry, len(entries))

	for _, entry := range entries {
		if err := queueEvent(entry); err != nil {
			return err
		}
	}

	return nil
}

func notifyMissedEvents() {
	str := strings.Join(func() []string {
		ret := make([]string, 0, len(missedEvents))
		for _, entry := range missedEvents {
			ret = append(ret, entry.Title)
		}
		return ret
	}(), "\n")
	beeep.Notify(fmt.Sprintf("You missed %d reminders", len(missedEvents)), str, "")
	for _, entry := range missedEvents {
		reminder.RemoveReminder(entry.Id)
	}
}

func StartServer(port string) (<-chan error, error) {
	if err := startEventListeners(); err != nil {
		return nil, err
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		switch r.Method {
		case http.MethodGet: // to check if the server is running
			w.Write([]byte("pong"))
		case http.MethodPost: // to add a reminder
			// read the request body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// unmarshal the request body into the event
			var event reminder.Entry
			if err := json.Unmarshal(body, &event); err != nil {
				log.Println(err)
				http.Error(w, "invalid json body", http.StatusBadRequest)
				return
			}

			if when, err := event.GetTime(); err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			} else if when.Before(time.Now()) {
				return
			}
			if err := queueEvent(event); err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			} else if err := reminder.AddReminder(event); err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case http.MethodDelete: // to remove a reminder
			// parse the id of the entry to be removed
			var (
				id  uuid.UUID
				err error
			)
			if idQuery := r.URL.Query().Get("id"); idQuery == "" {
				http.Error(w, "missing query 'id'", http.StatusBadRequest)
				return
			} else if id, err = uuid.Parse(idQuery); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			// remove the event from the list
			if _, err := reminder.RemoveReminder(id); err != nil {
				log.Println(err)
				http.Error(w, "failed to remove the event", http.StatusInternalServerError)
				return
			}
			if _, ok := tickerChannels[id]; ok {
				tickerChannels[id] <- true // close the ticker goroutine
			}
		default:
			http.Error(w, "invalid request type", http.StatusBadRequest)
		}
	})

	notifyMissedEvents()

	// main goroutine that waits for the event timers
	finished := make(chan error, 2)
	go func() {
		for event := range eventChan {
			if err := beeep.Notify(event.Title, event.Msg, ""); err != nil {
				fmt.Fprintf(os.Stderr, "%s", err)
			}
			if _, err := reminder.RemoveReminder(event.Id); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to remove Reminder: %s", err)
			}
		}
	}()
	go func() {
		finished <- http.ListenAndServe(port, nil)
	}()

	return finished, nil
}
