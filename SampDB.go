package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Data is the structure that holds the data to be written or read
type Computer struct {
	MAC		string `json:"mac"`
	Name		string `json:"name"`
	IP		string `json:"ip"`
	Assignee	string `json:"assignee"`
	Description	string `json:"description"`
}

type Assignment struct {
	Key		string `json:"key"`
	Assignee	string `json:"assignee"`
}

var dataStore []Computer;

func addComputer(w http.ResponseWriter, r *http.Request) {
	var c Computer
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if c.MAC == "" || c.Name == "" || c.IP == "" {
		http.Error(w, "Missing mandatory property.", http.StatusBadRequest)
		return
	}
	if c.Assignee != "" && len(c.Assignee) != 3 {
		http.Error(w, "'assignee' field is restricted to 3-letter employee codes.", http.StatusBadRequest)
		return // TODO: Reconsider limitation for future-proofing
	}
	dataStore = append(dataStore, c)

	// TODO: Add admin notification check here

	w.WriteHeader(http.StatusCreated)
}

func getComputerByMAC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("mac")
	for _, c := range(dataStore) {
		if c.MAC == key {
			json.NewEncoder(w).Encode(c)
			return
		}
	}
	http.Error(w, "Key not found", http.StatusNotFound)
}

func getComputerByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("name")
	for _, c := range(dataStore) {
		if c.Name == key {
			json.NewEncoder(w).Encode(c)
			return
		}
	}
	http.Error(w, "Key not found", http.StatusNotFound)
}

func getComputerByIP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("ip")
	for _, c := range(dataStore) {
		if c.IP == key {
			json.NewEncoder(w).Encode(c)
			return
		}
	}
	http.Error(w, "Key not found", http.StatusNotFound)
}

func getComputersByAssignee(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var cl []Computer
	key := r.URL.Query().Get("assignee")
	for _, c := range(dataStore) {
		if c.Assignee == key {
			cl = append(cl, c)
		}
	}
	if len(cl) == 0 {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(cl)
}

func getComputers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	if len(dataStore) == 0 {
		http.Error(w, "No items found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(dataStore)
}

func getUnassignedComputers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var cl []Computer
	for _, c := range(dataStore) {
		if c.Assignee == "" {
			cl = append(cl, c)
		}
	}
	if len(cl) == 0 {
		http.Error(w, "No unassigned items", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(cl)
}

func assignComputerByMAC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var a Assignment
	err := json.NewDecoder(r.Body).Decode(&a)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if a.Key == "" {
		http.Error(w, "Missing mandatory property 'key'.", http.StatusBadRequest)
		return
	}
	if a.Assignee == "" {
		http.Error(w, "Missing mandatory property 'assignee'.", http.StatusBadRequest)
		return
	}
	if len(a.Assignee) != 3 {
		http.Error(w, "'assignee' field is restricted to 3-letter employee codes.", http.StatusBadRequest)
		return // TODO: Reconsider limitation for future-proofing
	}
	for n, c := range(dataStore) {
		if c.MAC == a.Key {
			dataStore[n].Assignee = a.Assignee
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	// TODO: Add admin notification check here

	http.Error(w, "Key not found", http.StatusNotFound)
}

func assignComputerByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var a Assignment
	err := json.NewDecoder(r.Body).Decode(&a)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if a.Key == "" {
		http.Error(w, "Missing mandatory property 'key'.", http.StatusBadRequest)
		return
	}
	if a.Assignee == "" {
		http.Error(w, "Missing mandatory property 'assignee'.", http.StatusBadRequest)
		return
	}
	if len(a.Assignee) != 3 {
		http.Error(w, "'assignee' field is restricted to 3-letter employee codes.", http.StatusBadRequest)
		return // TODO: Reconsider limitation for future-proofing
	}
	for n, c := range(dataStore) {
		if c.Name == a.Key {
			dataStore[n].Assignee = a.Assignee
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	// TODO: Add admin notification check here

	http.Error(w, "Key not found", http.StatusNotFound)
}

func assignComputerByIP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var a Assignment
	err := json.NewDecoder(r.Body).Decode(&a)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if a.Key == "" {
		http.Error(w, "Missing mandatory property 'key'.", http.StatusBadRequest)
		return
	}
	if a.Assignee == "" {
		http.Error(w, "Missing mandatory property 'assignee'.", http.StatusBadRequest)
		return
	}
	if len(a.Assignee) != 3 {
		http.Error(w, "'assignee' field is restricted to 3-letter employee codes.", http.StatusBadRequest)
		return // TODO: Reconsider limitation for future-proofing
	}
	for n, c := range(dataStore) {
		if c.IP == a.Key {
			dataStore[n].Assignee = a.Assignee
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	// TODO: Add admin notification check here

	http.Error(w, "Key not found", http.StatusNotFound)
}

func unassignComputerByMAC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("mac")
	for n, c := range(dataStore) {
		if c.MAC== key {
			dataStore[n].Assignee = ""
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	http.Error(w, "Key not found", http.StatusNotFound)
}

func unassignComputerByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("name")
	for n, c := range(dataStore) {
		if c.Name == key {
			dataStore[n].Assignee = ""
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	http.Error(w, "Key not found", http.StatusNotFound)
}

func unassignComputerByIP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("ip")
	for n, c := range(dataStore) {
		if c.IP == key {
			dataStore[n].Assignee = ""
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	http.Error(w, "Key not found", http.StatusNotFound)
}

func deleteComputerByMAC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("mac")
	if len(dataStore) > 0 {
		if len(dataStore) == 1 && dataStore[0].MAC == key {
			dataStore = nil
		} else {
			for n, c := range(dataStore) {
				if c.MAC == key {
					dataStore[n] = dataStore[len(dataStore)-1]
					dataStore = dataStore[:len(dataStore)-1]
					break
				}
			}
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Key not found", http.StatusNotFound)
}

func deleteComputerByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("name")
	if len(dataStore) > 0 {
		if len(dataStore) == 1 && dataStore[0].Name == key {
			dataStore = nil
		} else {
			for n, c := range(dataStore) {
				if c.Name == key {
					dataStore[n] = dataStore[len(dataStore)-1]
					dataStore = dataStore[:len(dataStore)-1]
					break
				}
			}
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Key not found", http.StatusNotFound)
}

func deleteComputerByIP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("ip")
	if len(dataStore) > 0 {
		if len(dataStore) == 1 {
			dataStore = nil
		} else {
			for n, c := range(dataStore) {
				if c.IP == key {
					dataStore[n] = dataStore[len(dataStore)-1]
					dataStore = dataStore[:len(dataStore)-1]
					break
				}
			}
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Key not found", http.StatusNotFound)
}

func main() {
	http.HandleFunc("/getComputerByMAC",		getComputerByMAC)
	http.HandleFunc("/getComputerByName",		getComputerByName)
	http.HandleFunc("/getComputerByIP",		getComputerByIP)
	http.HandleFunc("/getComputersByAssignee",	getComputersByAssignee)
	http.HandleFunc("/getComputers",		getComputers)
	http.HandleFunc("/getUnassignedComputers",	getUnassignedComputers)
	http.HandleFunc("/addComputer",			addComputer)
	http.HandleFunc("/assignComputerByMAC",		assignComputerByMAC)
	http.HandleFunc("/assignComputerByName",	assignComputerByName)
	http.HandleFunc("/assignComputerByIP",		assignComputerByIP)
	http.HandleFunc("/unassignComputerByMAC",	unassignComputerByMAC)
	http.HandleFunc("/unassignComputerByName",	unassignComputerByName)
	http.HandleFunc("/unassignComputerByIP",	unassignComputerByIP)
	http.HandleFunc("/deleteComputerByMAC",		deleteComputerByMAC)
	http.HandleFunc("/deleteComputerByName",	deleteComputerByName)
	http.HandleFunc("/deleteComputerByIP",		deleteComputerByIP)
	fmt.Println("Starting server on port 55555...")
	http.ListenAndServe(":55555", nil)
	fmt.Println("Couldn't get a lock on the port. Is SampDB already running?")
}

