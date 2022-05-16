// package server handles the events, notifications
// and listens for incoming calls or updates
package server

import (
	"net/http"
)

func StartServer(port string) (<-chan error, error) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Write([]byte("Hello World"))
		case http.MethodPost:
		case http.MethodDelete:
		default:
			http.Error(w, "invalid request type", http.StatusBadRequest)
		}
	})

	finished := make(chan error)
	go func() {
		finished <- http.ListenAndServe(port, nil)
	}()

	return finished, nil
}
