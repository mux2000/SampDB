package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"net/http"
)

type Notification struct {
	Level		string `json:"level"`
	Employee	string `json:"employeeAbbreviation"`
	Message		string `json:"message"`
}

func notify(w http.ResponseWriter, r *http.Request) {
	var n Notification
	if r.Method != http.MethodPost {
		fmt.Printf("DummyListener: Wrong method\n")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&n)
	if err != nil {
		fmt.Printf("DummyListener: Wrong JSON object format\b")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	msgHeader := strings.ToUpper(n.Level)
	if msgHeader != "INFO" && msgHeader != "WARNING" && msgHeader != "ERROR" {
		fmt.Printf("DummyListener: Unexpected message level %s\n", n.Level)
		http.Error(w, "Unknown message type", http.StatusBadRequest)
	}

	if n.Employee != "" {
		msgHeader += " [" + n.Employee + "]: "
	} else {
		msgHeader += ": "
	}

	fmt.Printf("DummyListener %s: %s\n", msgHeader, n.Message)

	w.WriteHeader(http.StatusCreated)
}

func main() {
	http.HandleFunc("/api/notify", notify)

	fmt.Println("Starting server on port 8080...")
	http.ListenAndServe(":8080", nil)
}

