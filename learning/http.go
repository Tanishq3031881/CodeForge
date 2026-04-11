package main

import (
	"fmt"
	"net/http"
)

// handler for "/"
func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello!")
}

// handler for "/health"
func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

// handler for "/echo"
func echoHandler(w http.ResponseWriter, r *http.Request) {
	// get query param ?msg=
	msg := r.URL.Query().Get("msg")

	if msg == "" {
		fmt.Fprintln(w, "No message provided")
		return
	}

	fmt.Fprintln(w, msg)
}

func main() {
	// register routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/echo", echoHandler)

	fmt.Println("Server running on http://localhost:8080")

	// start server
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Server error:", err)
	}
}