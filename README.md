# SampDB - Sample Company Computer Database

## Introduction

The Sample Company Computer Database is a web server keeping track of the computers in Sample Company and their assignments to the different employees. The software allows administrators to add and delete computers from the database, assign, reassign and unassign computers from employees, and find out which computer is assigned to whom, which computers are assigned to a specific employee or which computers are currently unassigned, all through a convenient REST API.

### Preconditions

In order to build SampDB, the build machine must have the development package for sqlite3 installed. This can be accomplished using the following command:

   $ sudo apt-get install libsqlite3-dev

## Building the software

The software itself is contained in the sub-folder SampDB, but it also comes with a dummy listener device used for testing SampDB for issuing over-assignment notifications.

To build the dummy listener issue the following commands:

   $ cd DummyListener
   $ go build

To build SampDB issue the collowing commands:

   $ cd SampDB
   $ go get
   $ go build

## Testing the software

SampDB comes with a full set of tests for every aspect of the software. The tests themselves can be found in the file ```SampDB/SampDB_test.go```. To run the entire test suite, run the following commands:

   $ cd SampDB
   $ go get
   $ go test

To issue a specific test, issue the following command instead of the last one:

   $ go test -r <testname>

## Running the software

To run the software, once it's built, run the following command:

   $ ./SampDB/SampDB [--file <file>] --storage-type <volatile|json|sqlite>

The property **--file** is the name of the file to use for non-volatile data storage. This may be an SQLite or a JSON file depending on the choice of storage type. In case no file name is specified, the software will use default.json for JSON data and default.sqlite for SQLite formatted data.

The property **--storage-type** is the type of database SampDB will use to store the computers and their assignments.

 * **volatile** will keep the data in memory, meaning the database will be erased once the server is killed. This option is not recommended and is used for testing only. 
 * **json** will use a JSON formatted text file. This is a simple system that keeps the data in a human-readable format, making it easy to debug.
 * **sqlite** will use the SQLite database format. This is a highly efficient format used for high performance.

## Running the DummyListener service
To run the dummy listener service in order to test the communication with the notificationservice, run:

   $ ./DummyListener/DummyListener

## Communicating with the server

The following is a description of the REST API used by SampDB. This can be used to write scripts or software to interface with the database. All these endpoints are available at port 55555. If accessing the server locally, the URL will always be from one of the following forms:

   http://localhost:55555/<endpoint>

   http://localhost:55555/<endpoint>&paramerter=value


### Reading from the database

SampDB provides a few mechanisms for information retrieval. All use the **GET** HTTP method, and all can be accessed using the URL:

* getComputerByMAC allows the client to provide a MAC address by appending '&mac=<MAC>' to the end of the URL. The server will respond with a JSON object containing the computer with the specified MAC.
* getComputerByName allows the client to provide a computer name by appending '&name=<Name>' to the end of the URL. The server will respond with a JSON object containing the computer with the specified name.
* getComputerByIP allows the client to provide an IP address by appending '&ip=<IP>' to the end of the URL. The server will respond with a JSON object containing the computer with the specified IP address.
* getComputers will always respond with a JSON object containing all the computers.
* getComputersByAssignee allows the client to provide an assignee (3-letter code) by appending '&assignee=<assignee>' to the end of the URL. The server will respond with a JSON obect containing all the computers assigned to this employee.
* genUnassignedComputers will always respond with a JSON object containing all the unassigned computers.

### Adding computers to the database

SampDB provides one endpoint for adding new computers. **addCompueter** allows the user to add a new computer to the database. This is done by using a **POST** method call, with a body that is a JSON object containing the following fields:
1. MAC (mandatory, MAC address)
2. Name (mandatory, no spaces)
3. IP (mandatory, IP address)
4. Assignee (optional, employee code, 3 characters long)
5. Description (optional)

### Removing items from the database

SampDB provides three ways to specify a computer for deletion. These all use the **DELETE**HTTP method:

* deleteComputerByMAC allows the client to provide a MAC address by appending '&mac=<MAC>' to the end of the URL.
* deleteComputerByName allows the client to provide a computer name by appending 'name=<name>' to the end of the URL.
* deleteComputerByIP allows the client to provide an IP address by appending '&IP=<ip>' to the end of the URL.

### Assigning, Re-assigning and Unassigning computers to employees

SampDB provides several endpoints for the purpose of managing computer assignments. Assignement endpoints use the **PUT** HTTP method, while the unassignment endpoints use the **DELETE** HTTP method.

* assignComputerByMAC allows the client to provide an assignment within a JSON object containing the following fields:
1. Key - MAC address.
2. Assignee - 3 letter employee code.

* assignComputerByName allows the client to provide an assignment within a JSON object containing the following fields:
1. Key - Computer name.
2. Assignee - 3 letter employee code.

*assignComputerByIP allows the client to provide an assignment within a JSON object containing the following fields:
1. Key - IP address.
2. Assignee - 3 letter employee code.

All three assignement endpoints will instead reassign the computer if it is already assigned.

* UnassignComputerByMAC allows the client to specify the assignment to remove by appending '&mac=<MAC>' to the end of the URL.
* UnassignComputerByName allows the client to specify the assignment to remove by appending '&name=<Name>' to the end of the URL.
* UnassignComputerByIP allows the client to specify the assignment to remove by appending '&ip=<IP>' to the end of the URL.

All three unassignment endpoints do nothing if the computer is not already assigned.

## Overassignment notification service

In any event (either computer addition or computer assignment) that results in one employee being assigned three or more computers, SampDB will attempt to notify that fact to the system administrator. In order to do that, it will send a message to the address 'http://localhost:8080/api/notify.
