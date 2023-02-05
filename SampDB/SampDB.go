package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"bytes"
	"os"
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

type Notification struct {
        Level           string `json:"level"`
        Employee        string `json:"employeeAbbreviation"`
        Message         string `json:"message"`
}

func notify(emp string, numAssigned int) int {

	fmt.Printf("Warning: Employee [%s] has been assigned %d computers!\n", emp, numAssigned)

	var n = Notification {
		"Warning",
		emp,
		fmt.Sprintf("Over-assignement warning: Employee %s is now assigned %d computers.", emp, numAssigned),
	}

	jsonData, err := json.Marshal(n)
        if err != nil {
		fmt.Fprintf(os.Stderr, "Notification service: Error marshalling JSON object.")
                return -1
	}

	const listenerURL = "http://localhost:8080/api/notify"
	resp, err := http.Post(listenerURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Notification service: Error sending object (is the listener running?\n")
		return -1
	}
	defer resp.Body.Close()

	return resp.StatusCode
}

func checkEmployee (emp string) int {

	dataAccess.Lock()
	err, cl := dataStore.ReadAll(KeyAssignee, emp)
	dataAccess.Unlock()

	if err == errNotFound {
		return 0
	} else if err != nil {
		return http.StatusInternalServerError
	}

	if len(cl) > 2 {
		resp := notify(emp, len(cl))
		if resp < 0 {
			return resp
		} else if resp != http.StatusCreated {
			fmt.Printf("Notification returned %d\n", resp)
			return -1
		} else {
			return 0
		}
	}
	return 0
}

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

	dataAccess.Lock()
	err = dataStore.Add(c)
	dataAccess.Unlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if c.Assignee != "" {
		if checkEmployee(c.Assignee) < 0 {
			http.Error(w, "Error reporting over-assignement.", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func getComputerByMAC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("mac")

	dataAccess.Lock()
	err, c := dataStore.Read(KeyMAC, key)
	dataAccess.Unlock()

	if err == errNotFound {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(*c)
}

func getComputerByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("name")

	dataAccess.Lock()
	err, c := dataStore.Read(KeyName, key)
	dataAccess.Unlock()

	if err == errNotFound {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(*c)
}

func getComputerByIP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("ip")

	dataAccess.Lock()
	err, c := dataStore.Read(KeyIP, key)
	dataAccess.Unlock()

	if err == errNotFound {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(*c)
}

func getComputersByAssignee(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var cl []Computer
	key := r.URL.Query().Get("assignee")

	dataAccess.Lock()
	err, cl := dataStore.ReadAll(KeyAssignee, key)
	dataAccess.Unlock()

	if err == errNotFound {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(cl)
}

func getComputers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	dataAccess.Lock()
	err, cl := dataStore.ReadAll(KeyAll, "")
	dataAccess.Unlock()

	if err == errNotFound {
		http.Error(w, "No items found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(cl)
}

func getUnassignedComputers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	dataAccess.Lock()
	err, cl := dataStore.ReadAll(KeyNotAssigned, "")
	dataAccess.Unlock()

	if err == errNotFound {
		http.Error(w, "No unassigned items", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	dataAccess.Lock()
	err = dataStore.Assign(KeyMAC, a.Key, a.Assignee)
	dataAccess.Unlock()

	if err == errNotFound {
		http.Error(w, "Error items not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if checkEmployee(a.Assignee) < 0 {
		fmt.Printf("Error reporting over-assignement.")
		http.Error(w, "Error reporting over-assignement", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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

	dataAccess.Lock()
	err = dataStore.Assign(KeyName, a.Key, a.Assignee)
	dataAccess.Unlock()

	if err == errNotFound {
		http.Error(w, "Error items not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if checkEmployee(a.Assignee) < 0 {
		fmt.Printf("Error reporting over-assignement.")
		http.Error(w, "Error reporting over-assignement", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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

	dataAccess.Lock()
	err = dataStore.Assign(KeyIP, a.Key, a.Assignee)
	dataAccess.Unlock()

	if err == errNotFound {
		http.Error(w, "Error items not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if checkEmployee(a.Assignee) < 0 {
		fmt.Printf("Error reporting over-assignement.")
		http.Error(w, "Error reporting over-assignement", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func unassignComputerByMAC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("mac")

	dataAccess.Lock()
	err := dataStore.Unassign(KeyMAC, key)
	dataAccess.Unlock()

	if err == errNotFound {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func unassignComputerByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("name")

	dataAccess.Lock()
	err := dataStore.Unassign(KeyName, key)
	dataAccess.Unlock()

	if err == errNotFound {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func unassignComputerByIP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("ip")

	dataAccess.Lock()
	err := dataStore.Unassign(KeyName, key)
	dataAccess.Unlock()

	if err == errNotFound {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteComputerByMAC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("mac")

	dataAccess.Lock()
	err := dataStore.Delete(KeyMAC, key)
	dataAccess.Unlock()

	if err == errNotFound {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteComputerByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("name")

	dataAccess.Lock()
	err := dataStore.Delete(KeyName, key)
	dataAccess.Unlock()

	if err == errNotFound {
		http.Error(w, "Key not found", http.StatusNotFound)
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteComputerByIP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("ip")

	dataAccess.Lock()
	err := dataStore.Delete(KeyIP, key)
	dataAccess.Unlock()

	if err == errNotFound {
		http.Error(w, "Key not found", http.StatusNotFound)
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

var dataStore dataInterface

func main() {
	storagetype := flag.String("storage-type", "", "the type of storage to use ('volatile', 'json' or 'sqlite'")
	file := flag.String("file", "", "Optional. The file to use as database")
	flag.Parse()

	if *storagetype == "" ||
	   (*storagetype != "volatile" &&
	    *storagetype != "json" &&
	    *storagetype != "sqlite"){
		fmt.Println("Usage: SampDB [--file=<file>] --storage-type=<volatile|json|sqlite>")
		return
	}

	if *storagetype == "json" && *file == "" {
		*file = "default.json"
	} else if *storagetype == "sqlite" && *file == "" {
		*file = "default.sqlite"
	}

	err := GetDataStore(*storagetype, *file, &dataStore)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %s\n", err.Error())
		return
	}

	if dataStore == nil {
		fmt.Fprintf(os.Stderr, "Error initializing database.\n")
		return
	}

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

