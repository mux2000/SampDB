package main

import (
	"testing"
	"strings"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

const (
	errMarshalling int = -1
	errUnmarshalling   = -2
	errSending         = -3
	errReceiving       = -4
	errReading         = -5
	errFailed          = -6
)

const baseURL = "http://localhost:55555"

func handleError(t *testing.T, err int, req string) {
	if err == errMarshalling {
		t.Errorf("Error marshalling for %s.", req)
	} else if err == errUnmarshalling {
		t.Errorf("Error unmarshaliing response for %s.", req)
	} else if err == errSending {
		t.Errorf("Error sending request %s.", req)
	} else if err == errReceiving {
		t.Errorf("Error receiving response for %s.", req)
	} else if err == errReading {
		t.Errorf("Error reading body of response for %s.", req)
	} else if err == errFailed {
		t.Errorf("Unexpected error performing %s.", req)
	} else {
		t.Errorf("Unexpected response code %d revceived for request %s.", err, req)
	}
}

var outSampDBBuf bytes.Buffer
var outDummyListenerBuf bytes.Buffer
var SampDBBuilt = false
var DummyListenerBuilt = false
var SampDB, DummyListener *exec.Cmd

func init () {

	// Kill servers if they're already running.
	exec.Command("killall SampDB").Output()
	exec.Command("killall DummyListener").Output()
}

func setupTest (t *testing.T) {

	var err error

	fmt.Printf("Setting up test...\n")

	// Build SampDB if it isn't already built
	if !SampDBBuilt {
		buildCmd := exec.Command("go", "build")
		err = buildCmd.Run()
		if err != nil {
			t.Fatalf("Error building SampDB: %s", err.Error())
			return
		}
		SampDBBuilt = true
	}

	// Run SampDB in the background
	outSampDBBuf.Reset()
	SampDB = exec.Command("./SampDB")
        SampDB.Stdout = &outSampDBBuf
	SampDB.Stderr = os.Stderr
	err = SampDB.Start()
	if err != nil {
		t.Fatalf("Error starting program: %s", err.Error())
		return
	}

	// Verify server is up and running
	timeout := 10
	startLine := "Starting server on port 55555...\n"
	for timeout > 0 {
		if outSampDBBuf.String() == startLine {
			break
		}
		time.Sleep(time.Second)
		timeout --
		if timeout == 0 {
			t.Fatalf("Error starting SampDB - timeout reached.")
		}
	}
	fmt.Printf(outSampDBBuf.String())

	err = os.Chdir("../DummyListener")
	if err != nil {
		t.Fatalf("Error changing to DummyListener folder. Is this executed in the right folder?")
	}

	if !DummyListenerBuilt {
		// Build DummyListener if it isn't already built
		buildCmd := exec.Command("go", "build")
		err = buildCmd.Run()
		if err != nil {
			t.Fatalf("Error building DummyListener: %s", err.Error())
			return
		}
		DummyListenerBuilt = true
	}

	// Run DummyListener in the background
	outDummyListenerBuf.Reset()
	DummyListener = exec.Command("./DummyListener")
        DummyListener.Stdout = &outDummyListenerBuf
	err = DummyListener.Start()
	if err != nil {
		t.Fatalf("Error starting program: %s", err.Error())
		return
	}

	// Verify server is up and running
	timeout = 10
	startLine = "Starting server on port 8080...\n"
	for timeout > 0 {
		if outDummyListenerBuf.String() == startLine {
			break
		}
		time.Sleep(time.Second)
		timeout --
		if timeout == 0 {
			t.Fatalf("Error starting SampDB - timeout reached.")
		}
	}
	fmt.Printf(outDummyListenerBuf.String())

	err = os.Chdir("../SampDB")
	if err != nil {
		t.Fatalf("Error changing to SampDB folder. Is this executed in the right folder?")
	}

	fmt.Printf("Setup complete.\n")
}

func teardownTest (t *testing.T) {

	fmt.Printf("Tearing down test...\n")
	// Kill SampDB
	err := SampDB.Process.Kill()
	if err != nil {
		t.Fatalf("Error killing SampDB: %s", err.Error())
		return
	}
	SampDB.Wait()
	fmt.Printf("SampDB process terminated.\n")

	// Kill DummyListener
	err = DummyListener.Process.Kill()
	if err != nil {
		t.Fatalf("Error killing DummyListener: %s", err.Error())
		return
	}
	DummyListener.Wait()
	fmt.Printf("DummyListener process terminated.\n")

	fmt.Printf("Teardown complete.\n")
}

func addComputerReq(t *testing.T, c Computer) int {
	fmt.Printf("Adding computer %v\n", c)
        jsonData, err := json.Marshal(c)
	if err != nil {
		return errMarshalling
	}

	resp, err := http.Post(baseURL + "/addComputer", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return errSending
	}
	defer resp.Body.Close()

	return resp.StatusCode
}

func getComputerByReq(t *testing.T, keyname, key string) (int, Computer) {
	fmt.Printf("Getting computer with %s=%s\n", keyname, key)
	var c Computer
	resp, err := http.Get(fmt.Sprintf("%s/getComputerBy%s?%s=%s", baseURL, keyname, strings.ToLower(keyname), key))
	if err != nil {
		return errSending, c
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errReading, c
		}

		err = json.Unmarshal(body, &c)
		if err != nil {
			return errUnmarshalling, c
		}
		return resp.StatusCode, c
	}
	return resp.StatusCode, c
}

func getComputersByAssigneeReq(t *testing.T, assignee string) (int, []Computer) {
	fmt.Printf("Getting computers with assignee=%s\n", assignee)
	var cl []Computer
	resp, err := http.Get(fmt.Sprintf("%s/getComputersByAssignee?assignee=%s", baseURL, assignee))
	if err != nil {
		return errSending, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errReading, nil
		}

		err = json.Unmarshal(body, &cl)
		if err != nil {
			return errUnmarshalling, nil
		}
		return resp.StatusCode, cl
	}
	return resp.StatusCode, nil
}

func getUnassignedComputersReq(t *testing.T) (int, []Computer) {
	fmt.Printf("Getting all unassigned computers.\n")
	var cl []Computer
        resp, err := http.Get(fmt.Sprintf("%s/getUnassignedComputers", baseURL))
        if err != nil {
                return errSending, nil
        }
        defer resp.Body.Close()

        if resp.StatusCode == http.StatusOK {

                body, err := ioutil.ReadAll(resp.Body)
                if err != nil {
                        return errReading, nil
                }

                err = json.Unmarshal(body, &cl)
                if err != nil {
                        return errUnmarshalling, nil
                }
                return resp.StatusCode, cl
	}
        return resp.StatusCode, nil
}

func getComputersReq(t *testing.T) (int, []Computer) {
	fmt.Printf("Getting all computers.\n")
	var cl []Computer
	resp, err := http.Get(fmt.Sprintf("%s/getComputers", baseURL))
        if err != nil {
                return errSending, nil
        }
        defer resp.Body.Close()

        if resp.StatusCode == http.StatusOK {

                body, err := ioutil.ReadAll(resp.Body)
                if err != nil {
                        return errReading, nil
                }

                err = json.Unmarshal(body, &cl)
                if err != nil {
                        return errUnmarshalling, nil
                }
                return resp.StatusCode, cl
	}
        return resp.StatusCode, nil
}

func assignComputerByReq(t *testing.T, keyname, key, assignee string) int {
	fmt.Printf("Assigning computer with %s=%s to %s\n", keyname, key, assignee)
	var a = Assignment { key, assignee }

	jsonData, err := json.Marshal(a)
	if err != nil {
		return errMarshalling
	}

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/assignComputerBy%s?%s=%s&Assignee=%s", baseURL, keyname, strings.ToLower(keyname), key, assignee), bytes.NewBuffer(jsonData))
	if err != nil {
		return errSending
	}
	resp, err := client.Do(req)
	if err != nil {
		return errSending
	}

	return resp.StatusCode
}

func unassignComputerByReq(t *testing.T, keyname, key string) int {
	fmt.Printf("Unassigning computer with %s=%s\n", keyname, key)
	client := &http.Client{}
        req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/unassignComputerBy%s?%s=%s", baseURL, keyname, strings.ToLower(keyname), key), nil)
        if err != nil {
                return errSending
        }
        resp, err := client.Do(req)
	if err != nil {
		return errSending
	}

	return resp.StatusCode
}

func delComputerByReq(t *testing.T, keyname, key string) int {
	fmt.Printf("Deleting computer with %s=%s\n", keyname, key)
        client := &http.Client{}
        req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/deleteComputerBy%s?%s=%s", baseURL, keyname, strings.ToLower(keyname), key), nil)
        if err != nil {
                return errSending
        }
        resp, err := client.Do(req)
	if err != nil {
		return errSending
	}

	return resp.StatusCode
}

func TestDeleteAllComputers(t *testing.T) {

	fmt.Printf("Starting test 'DeleteAllComputers.\n")

	setupTest(t)

	for i := 0; i <= 9; i++ {
		c := Computer {
			MAC: fmt.Sprintf("0%d:2%d:4%d:6%d:8%d:a%d", i, i, i, i, i ,i),
			Name: fmt.Sprintf("TestComputer%d", i),
			IP: fmt.Sprintf("172.1.0.%d", i),
			Assignee: "",
			Description: "",
		}
	        // Add computer
		resp := addComputerReq(t, c)
		if resp != http.StatusCreated {
			handleError(t, resp, "addComputer")
		}
	}

	resp, cl := getComputersReq(t)
	if resp != http.StatusOK {
		handleError(t, resp, "getComputersReq")
        }

	if len(cl) != 10 {
		t.Fatalf("Unexpected number of computers read. Expected 10 but got %d", len(cl))
	}

	for n, c := range(cl) {
		if n % 3 == 0 {
			resp := delComputerByReq(t, "MAC", c.MAC)
			if resp != http.StatusOK {
				handleError(t, resp, "deleteComputersByMAC")
			}
		} else if n % 3 == 1 {
			resp := delComputerByReq(t, "Name", c.Name)
			if resp != http.StatusOK {
				handleError(t, resp, "deleteComputersByName")
			}
		} else {
			resp := delComputerByReq(t, "IP", c.IP)
			if resp != http.StatusOK {
				handleError(t, resp, "deleteComputersByIP")
			}
		}
	}

	resp, cl = getComputersReq(t)
	if resp == http.StatusOK {
		t.Errorf("Unexpected response %d for getComputers (expected StatusNotFound).", resp)
	}
}

	teardownTest(t)

	fmt.Printf("Test 'DeleteAllComputers' completed.\n");
}


func TestAddReadRemoveMinimal(t *testing.T) {

	fmt.Printf("Starting test 'AddReadRemoveMininal'\n")

	setupTest(t)

	var c = Computer {
		MAC: "01:23:45:67:89:ab",
		Name: "TestComputer",
		IP: "172.1.0.1",
		Assignee: "",
		Description: "",
	}

	// Add computer
	resp := addComputerReq(t, c)
	if resp != http.StatusCreated {
		handleError(t, resp, "addComputer")
	}

	// Try to get it back in 3 different ways
	resp, got := getComputerByReq(t, "MAC", c.MAC)
	if resp != http.StatusOK {
		handleError(t, resp, "getComputerByMAC")
	}
	if got != c {
		t.Errorf("MAC get failed. Read item is different to the one added: %v != %v", got, c)
	}

	resp, got = getComputerByReq(t, "Name", c.Name)
	if resp != http.StatusOK {
		handleError(t, resp, "getComputerByName")
	}
	if got != c {
		t.Errorf("Name get failed. Read item is different to the one added: %v != %v", got, c)
	}

	resp, got = getComputerByReq(t, "IP", c.IP)
	if resp != http.StatusOK {
		handleError(t, resp, "getComputerByIP")
	}
	if got != c {
		t.Errorf("IP get failed. Read item is different to the one added: %v != %v", got, c)
	}

	// Delete the computer to return to baseline.
	resp = delComputerByReq(t, "MAC", c.MAC)
	if resp != http.StatusOK {
		handleError(t, resp, "deleteComputerByMAC")
	}

	teardownTest(t)

	fmt.Printf("Test 'AddReadRemoveMininal' complete.\n")
}

func TestAddMalformed(t *testing.T) {

	fmt.Printf("Starting test 'Malformed'.\n")

	setupTest(t)

	var malformed = Computer {
		MAC: "01:23:45:67:89:ab",
		Name: "Malformed",
		IP: "172.1.0.1",
		Assignee: "",
		Description: "",
	}

	// Try to add a computer without a MAC.
	malformed.MAC = ""
	resp := addComputerReq(t, malformed)
	if resp != http.StatusBadRequest {
		t.Errorf("Error %d received instead of StatusBadRequest in malformed addComputer request.", resp)
	}

	// Try to add a computer without a name.
	malformed.MAC = "01:23:45:67:89:ab"
	malformed.Name = ""
	resp = addComputerReq(t, malformed)
	if resp != http.StatusBadRequest {
		t.Errorf("Error %d received instead of StatusBadRequest in malformed addComputer request.", resp)
	}

	// Try to add a computer without an IP.
	malformed.Name = "Malformed"
	malformed.IP = ""
	resp = addComputerReq(t, malformed)
	if resp != http.StatusBadRequest {
		t.Errorf("Error %d received instead of StatusBadRequest in malformed addComputer request.", resp)
	}

	// Try to add a computer with a short assigned employee code
	malformed.IP = "172.1.0.1"
	malformed.Assignee = "ab"
	resp = addComputerReq(t, malformed)
	if resp != http.StatusBadRequest {
		t.Errorf("Error %d received instead of StatusBadRequest in malformed addComputer request.", resp)
	}

	// Try to add a computer with a long assigned employee code
	malformed.Assignee = "abcd"
	resp = addComputerReq(t, malformed)
	if resp != http.StatusBadRequest {
		t.Errorf("Error %d received instead of StatusBadRequest in malformed addComputer request.", resp)
	}

	// Try to assign a computer to a short employee code
	malformed.Assignee = "abc"
	resp = addComputerReq(t, malformed)
	if resp != http.StatusCreated {
		handleError(t, resp, "addComputer")
	}

	resp = assignComputerByReq(t, "MAC", malformed.MAC, "ab")
	if resp != http.StatusBadRequest {
		t.Errorf("Error %d received instead of StatusBadRequest in malformed assignComputerByMAC request.", resp)
	}
	resp = assignComputerByReq(t, "Name", malformed.Name, "ab")
	if resp != http.StatusBadRequest {
		t.Errorf("Error %d received instead of StatusBadRequest in malformed assignComputerByName request.", resp)
	}
	resp = assignComputerByReq(t, "IP", malformed.IP, "ab")
	if resp != http.StatusBadRequest {
		t.Errorf("Error %d received instead of StatusBadRequest in malformed assignComputerByIP request.", resp)
	}

	// Try to assign a computer to a long employee code
	resp = assignComputerByReq(t, "MAC", malformed.MAC, "abcd")
	if resp != http.StatusBadRequest {
		t.Errorf("Error %d received instead of StatusBadRequest in malformed assignComputerByMAC request.", resp)
	}
	resp = assignComputerByReq(t, "Name", malformed.Name, "abcd")
	if resp != http.StatusBadRequest {
		t.Errorf("Error %d received instead of StatusBadRequest in malformed assignComputerByName request.", resp)
	}
	resp = assignComputerByReq(t, "IP", malformed.IP, "abcd")
	if resp != http.StatusBadRequest {
		t.Errorf("Error %d received instead of StatusBadRequest in malformed assignComputerByIP request.", resp)
	}

	// Delete the computer to return to baseline.
	resp = delComputerByReq(t, "Name", malformed.Name)
	if resp != http.StatusOK {
		handleError(t, resp, "deleteComputerByName")
	}

	teardownTest(t)

	fmt.Printf("Test 'Malformed' complete.\n");
}

func TestAssignUnassign(t *testing.T) {

	fmt.Printf("Starting test 'AssignUnassign'.\n")

	setupTest(t)

	var cl []Computer
	var c1 = Computer {
		MAC: "01:23:45:67:89:ab",
		Name: "TestComputer1",
		IP: "172.1.0.1",
		Assignee: "mmu",
		Description: "Test description 1",
	}
	var c2 = Computer {
		MAC: "cd:ef:ba:ad:ca:fe",
		Name: "TestComputer2",
		IP: "172.1.0.2",
		Assignee: "",
		Description: "Test description 2",
	}

	// Add assigned computer
	resp := addComputerReq(t, c1)
	if resp != http.StatusCreated {
		handleError(t, resp, "addComputer")
	}

	// Add unassigned computer
	resp = addComputerReq(t, c2)
	if resp != http.StatusCreated {
		handleError(t, resp, "addComputer")
	}

	// Verify only assigned computer is returned by getComputerByAssignee
	resp, cl = getComputersByAssigneeReq(t, "mmu")
	if resp != http.StatusOK {
		handleError(t, resp, "getComputerByAssignee")
	}
	if len(cl) != 1 {
		t.Errorf("Unexpected number of items returned by getComputersByAssignee (expected 1, got %d).", len(cl))
	}
	if len(cl) > 0 && cl[0] != c1 {
		t.Errorf("Unexpected item returned by getComputersByAssignee(%v != %v).", cl[0], c1)
	}

	// Verify only unassigned computer is returned by getUnassignedComputers
	resp, cl = getUnassignedComputersReq(t)
	if resp != http.StatusOK {
		handleError(t, resp, "getUnassignedComputers")
	}
	if len(cl) != 1 {
	        t.Errorf("Unexpected number of items returned by getComputersByAssignee (expected 1, got %d).", len(cl))
        }
	if len(cl) > 0 && cl[0] != c2 {
	        t.Errorf("Unexpected item returned by getComputersByAssignee.")
	}

	// Assign the unassigned computer
	resp = assignComputerByReq(t, "MAC", c2.MAC, "mmu")
	if resp != http.StatusOK {
		handleError(t, resp, "assignComputerByMAC")
	}
	c2.Assignee = "mmu"

        // Verify both computers are returned by getComputerByAssignee
	resp, cl = getComputersByAssigneeReq(t, "mmu")
	if resp != http.StatusOK {
		handleError(t, resp, "getComputerByAssignee")
	}
        if len(cl) != 2 {
		t.Errorf("Unexpected number of items returned by getComputersByAssignee (expected 1, got %d).", len(cl))
	}
	if len(cl) > 1 {
		if !((cl[0].MAC == c1.MAC && cl[1].MAC == c2.MAC) ||
		(cl[0].MAC == c2.MAC && cl[1].MAC == c1.MAC)){
			t.Errorf("Unexpected items returned by getComputersByAssignee.")
		}
	}

	// Verify no computers are returned by getUnassignedComputers
	resp, cl = getUnassignedComputersReq(t)
	if resp != http.StatusNotFound {
		t.Errorf("Unexpected response %d for getUnassignedComputers (expected StatusNotFound).", resp)
	}

	// Unassign the first computer
	resp = unassignComputerByReq(t, "MAC", c1.MAC)
	if resp != http.StatusOK {
		handleError(t, resp, "unassignComputerByIP")
	}
	c1.Assignee = ""

	// Verify only the second computer is returned by getComputerByAssignee
	resp, cl = getComputersByAssigneeReq(t, "mmu")
	if resp != http.StatusOK{
		handleError(t, resp, "getComputerByAssignee")
	}
	if len(cl) != 1 {
		t.Errorf("Unexpected number of items returned by getComputersByAssignee (expected 1, got %d).", len(cl))
	}
	if len(cl) > 0 && cl[0] != c2 {
		t.Errorf("Unexpected item returned by getComputersByAssignee(%v != %v).", cl[0], c1)
	}

	// Verify only unassigned computer is returned by getUnassignedComputers
	resp, cl = getUnassignedComputersReq(t)
	if resp != http.StatusOK {
		handleError(t, resp, "getUnassignedComputers")
	}
	if len(cl) != 1 {
		t.Errorf("Unexpected number of items returned by getComputersByAssignee (expected 1, got %d).", len(cl))
	}
	if len(cl) > 0 && cl[0] != c1 {
		t.Errorf("Unexpected items returned by getComputersByAssignee.")
	}

	// Delete the computers to return to baseline.
	resp = delComputerByReq(t, "Name", c1.Name)
	if resp != http.StatusOK {
		handleError(t, resp, "deleteComputerByName")
	}
	resp = delComputerByReq(t, "IP", c2.IP)
	if resp != http.StatusOK {
		handleError(t, resp, "deleteComputerByIP")
	}

	teardownTest(t)

	fmt.Printf("Test 'AssignUnassign' complete.\n");
}

func TestReassign(t *testing.T) {

	fmt.Printf("Starting test 'Reassign'.\n");

	setupTest(t)

	var cl []Computer
	var c1 = Computer {
		MAC: "01:23:45:67:89:ab",
		Name: "TestComputer1",
		IP: "172.1.0.1",
		Assignee: "mmu",
		Description: "",
	}
	var c2 = Computer {
		MAC: "cd:ef:ba:ad:ca:fe",
		Name: "TestComputer2",
		IP: "172.1.0.2",
		Assignee: "ima",
		Description: "",
	}

	// Add computers
	resp := addComputerReq(t, c1)
	if resp != http.StatusCreated {
		handleError(t, resp, "addComputer")
	}
	resp = addComputerReq(t, c2)
	if resp != http.StatusCreated {
		handleError(t, resp, "addComputer")
	}

	// Verify only correctly assigned computer is returned by getComputerByAssignee
	resp, cl = getComputersByAssigneeReq(t, "mmu")
	if resp != http.StatusOK{
		handleError(t, resp, "getComputerByAssignee")
	}
	if len(cl) != 1 {
		t.Errorf("Unexpected number of items returned by getComputersByAssignee (expected 1, got %d).", len(cl))
	}
	if len(cl) > 0 && cl[0] != c1 {
		t.Errorf("Unexpected item returned by getComputersByAssignee(%v != %v).", cl[0], c1)
	}
	resp, cl = getComputersByAssigneeReq(t, "ima")
	if resp != http.StatusOK {
		handleError(t, resp, "getComputerByAssignee")
	}
	if len(cl) != 1 {
		t.Errorf("Unexpected number of items returned by getComputersByAssignee (expected 1, got %d).", len(cl))
	}
	if len(cl) > 0 && cl[0] != c2 {
		t.Errorf("Unexpected item returned by getComputersByAssignee(%v != %v).", cl[0], c2)
	}

	// Reassign computers
	resp = assignComputerByReq(t, "Name", c2.Name, "mmu")
	if resp != http.StatusOK {
		handleError(t, resp, "assignComputerByName")
	}
	c2.Assignee = "mmu"
	resp = assignComputerByReq(t, "IP", c1.IP, "ima")
	if resp != http.StatusOK{
		handleError(t, resp, "assignComputerByIP")
	}
	c1.Assignee = "ima"

	// Verify only correctly assigned computer is returned by getComputerByAssignee
	resp, cl = getComputersByAssigneeReq(t, "mmu")
	if resp != http.StatusOK {
		handleError(t, resp, "getComputerByAssignee")
	}
	if len(cl) != 1 {
		t.Errorf("Unexpected number of items returned by getComputersByAssignee (expected 1, got %d).", len(cl))
	}
	if len(cl) > 0 && cl[0] != c2 {
		t.Errorf("Unexpected item returned by getComputersByAssignee(%v != %v).", cl[0], c1)
	}
	resp, cl = getComputersByAssigneeReq(t, "ima")
	if resp != http.StatusOK {
		handleError(t, resp, "getComputerByAssignee")
	}
	if len(cl) != 1 {
		t.Errorf("Unexpected number of items returned by getComputersByAssignee (expected 1, got %d).", len(cl))
	}
	if len(cl) > 0 && cl[0] != c1 {
		t.Errorf("Unexpected item returned by getComputersByAssignee(%v != %v).", cl[0], c2)
	}

	// Delete the computers to return to baseline.
	resp = delComputerByReq(t, "Name", c1.Name)
	if resp != http.StatusOK {
		handleError(t, resp, "deleteComputerByName")
	}
	resp = delComputerByReq(t, "IP", c2.IP)
	if resp != http.StatusOK {
		handleError(t, resp, "deleteComputerByIP")
	}

	teardownTest(t)

	fmt.Printf("Test 'Reassign' complete.\n")
}

func TestNotification(t *testing.T) {

	fmt.Printf("Starting test 'Notification'\n")

	setupTest(t)

	// Add three computers, assigned to the same person
	for i := 0; i < 3; i ++ {
		resp := addComputerReq(t,  Computer{
			        fmt.Sprintf("0%d:2%d:4%d:6%d:8%d:a%d",i,i,i,i,i,i),
				fmt.Sprintf("TestComputer%d", i),
				fmt.Sprintf("172.1.0.%d", i),
				"mmu",
				"",
			})

	        if resp != http.StatusCreated {
		        handleError(t, resp, "addComputer")
		}
	}

	// Loooking for the warning from SampDB and the DummyListener
	expectedSampDB := `Starting server on port 55555...
Warning: Employee [mmu] has been assigned 3 computers!
`

	expectedDummyListener := `Starting server on port 8080...
DummyListener WARNING [mmu]: : Over-assignement warning: Employee mmu is now assigned 3 computers.
`

	if outSampDBBuf.String() != expectedSampDB {
		t.Errorf("Expected SampDB output:\n%q\nBut got: %s",
				expectedSampDB, outSampDBBuf.String())
	}
	if string(outDummyListenerBuf.Bytes()) != expectedDummyListener {
		t.Errorf("Expected dummyListener output:\n%q\nBut got: %s",
			expectedDummyListener, outSampDBBuf.String())
	}

	// Reset output buffers
	outSampDBBuf.Reset()
	outDummyListenerBuf.Reset()

	// Adding three non-assigned computers
	for i := 3; i < 6; i ++ {
		resp := addComputerReq(t,  Computer{
			        fmt.Sprintf("0%d:2%d:4%d:6%d:8%d:a%d",i,i,i,i,i,i),
				fmt.Sprintf("TestComputer%d", i),
				fmt.Sprintf("172.1.0.%d", i),
				"",
				"",
			})

	        if resp != http.StatusCreated {
		        handleError(t, resp, "addComputer")
		}
	}

	// Assigning computers to the same user
	resp := assignComputerByReq(t, "MAC", "03:23:43:63:83:a3", "ima")
	if resp != http.StatusOK {
		handleError(t, resp, "assignComputerByMAC")
	}
	resp = assignComputerByReq(t, "Name", "TestComputer4", "ima")
	if resp != http.StatusOK {
		handleError(t, resp, "assignComputerByName")
	}
	resp = assignComputerByReq(t, "IP", "172.1.0.5", "ima")
	if resp != http.StatusOK {
		handleError(t, resp, "assignComputerByIP")
	}

	// Loooking for the warnings from SampDB and the DummyListener
	expectedSampDB = `Starting server on port 55555...
Warning: Employee [mmu] has been assigned 3 computers!
Warning: Employee [ima] has been assigned 3 computers!
`

	expectedDummyListener = `Starting server on port 8080...
DummyListener WARNING [mmu]: : Over-assignement warning: Employee mmu is now assigned 3 computers.
DummyListener WARNING [ima]: : Over-assignement warning: Employee ima is now assigned 3 computers.
`


	if outSampDBBuf.String() != expectedSampDB {
		t.Errorf("Expected SampDB output:\n%q\nBut got: %s",
				expectedSampDB, outSampDBBuf.String())
	}
	if outDummyListenerBuf.String() != expectedDummyListener {
		t.Errorf("Expected dummyListener output:\n%q\nBut got: %s",
			expectedDummyListener, outDummyListenerBuf.String())
	}

	for i := 0; i < 6; i ++ {
		resp = delComputerByReq(t, "Name", fmt.Sprintf("TestComputer%d", i))
		if resp != http.StatusOK {
			handleError(t, resp, "deleteComputerByMAC")
		}
	}

	teardownTest(t)

	fmt.Printf("Test 'Notification' complete.\n")
}
