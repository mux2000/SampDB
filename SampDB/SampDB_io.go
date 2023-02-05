package main

import (
	"encoding/json"
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/gwenn/gosqlite"
)

var errUnknownDBType = errors.New("unkown database type")
var errCreatingDB = errors.New("error creating database")
var errOpeningDB = errors.New("error opening database")
var errReadingDB = errors.New("error reading from database")
var errWritingDB = errors.New("error writing to database")
var errClosingDB = errors.New("error closing database")
var errUnknownKeyType = errors.New("unknown key type")
var errInvalidKeyType = errors.New("Invalid key type for this operation")
var errNotFound = errors.New("record not found")
var errAlreadyExists = errors.New("record already exists")
var errNotUnique = errors.New("record is not uniquely described by key value pair")
var errMalformed = errors.New("malformed record")

const (
	KeyMAC		= "MAC"
	KeyName		= "Name"
	KeyIP		= "IP"
	KeyAll		= "All"
	KeyAssignee	= "Assignee"
	KeyNotAssigned	= "NotAssigned"
)

type dataInterface interface {
	Read (string, string) (error, *Computer)
	ReadAll (string, string) (error, []Computer)
	Add (Computer) error
	Delete (string, string) error
	Assign (string, string, string) error
	Unassign (string, string) error
	Close() error
}

func GetDataStore(dbtype, file string, db *dataInterface) error {
	if dbtype == "volatile" {
		return initVolatile(db)
	}
	if dbtype == "json" {
		return initJSON(db, file)
	}
	if dbtype == "sqlite" {
		return initSQL(db, file)
	}
	fmt.Fprintf(os.Stderr, "Error unknown DB type %s.\n", dbtype)
	return errUnknownDBType
}

/****************/
/* Non-Volatile */
/****************/

type volatileStore struct {
	data []Computer
}

var v volatileStore

func initVolatile (vp *dataInterface) error {
	*vp = (dataInterface)(&v)
	return nil
}

func (v *volatileStore) Read (keytype, key string) (error, *Computer) {
	if keytype != KeyMAC && keytype != KeyName && keytype != KeyIP {
		if keytype == KeyAssignee || keytype == KeyNotAssigned {
			fmt.Fprintf(os.Stderr, "Error fetching item: Invalid key type %s.\n", keytype)
			return errInvalidKeyType, nil
		}
		fmt.Fprintf(os.Stderr, "Error fetching item: Unknown key type %s.\n", keytype)
		return errUnknownKeyType, nil
	}
	var found *Computer = nil
	for _, c := range(v.data) {
		if (keytype == KeyMAC && c.MAC == key) ||
		   (keytype == KeyName && c.Name == key) ||
		   (keytype == KeyIP && c.IP == key) {
			if found != nil {
				fmt.Fprintf(os.Stderr, "Error fetching item with %s=%s: Multiple entries found.\n", keytype, key)
				return errNotUnique, nil
			}
			found = &c
		}
	}
	if found != nil {
		return nil, found
	}

	fmt.Fprintf(os.Stderr, "Error fetching item with %s=%s: No item found.\n", keytype, key)
	return errNotFound, nil
}

func (v *volatileStore) ReadAll (keytype, key string) (error, []Computer) {
	var cl []Computer

	if keytype == KeyAssignee {
		for _, c := range(v.data) {
			if c.Assignee == key {
				cl = append(cl, c)
			}
		}
	} else if keytype == KeyNotAssigned {
		for _, c := range(v.data) {
			if c.Assignee == "" {
				cl = append(cl, c)
			}
		}
	} else if keytype == KeyAll {
		for _, c := range(v.data) {
			cl = append(cl, c)
			fmt.Printf("@")
		}
	} else if keytype == KeyMAC || keytype == KeyName || keytype == KeyIP {
		fmt.Fprintf(os.Stderr, "Error fetching items: Invalid key type %s.\n", keytype)
		return errInvalidKeyType, nil
	} else {
		fmt.Fprintf(os.Stderr, "Error fetching items: Unknown key type %s.\n", keytype)
		return errUnknownKeyType, nil
	}

	if len(cl) > 0 {
		return nil, cl
	}

	fmt.Fprintf(os.Stderr, "Error fetching items with %s=%s: No items found.\n", keytype, key)
	return errNotFound, nil
}

func (v *volatileStore) Add (c Computer) error {
	if c.MAC == "" || c.Name == "" || c.IP == "" {
		fmt.Fprintf(os.Stderr, "Error adding item: MAC, Name and IP are mandatory fields.\n")
		return errMalformed
	}
	if c.Assignee != "" && len(c.Assignee) != 3 {
		fmt.Fprintf(os.Stderr, "Error adding item: Assignee code must be exactly three characters long.\n")
		return errMalformed
	}
	for _, nvc := range(v.data) {
		if c.MAC == nvc.MAC || c.Name == nvc.Name || c.IP == nvc.IP {
			fmt.Fprintf(os.Stderr, "Error adding item: Item already exists.\n")
			return errAlreadyExists
		}
	}
	v.data = append(v.data, c)

	return nil
}

func (v *volatileStore) Delete (keytype, key string) error {
	if keytype != KeyMAC && keytype != KeyName && keytype != KeyIP {
		if keytype == KeyAssignee || keytype == KeyNotAssigned {
			fmt.Fprintf(os.Stderr, "Error deleting item: Invalid key type %s.\n", keytype)
			return errInvalidKeyType
		}
		fmt.Fprintf(os.Stderr, "Error deleting item: Unknown key type %s.\n", keytype)
		return errUnknownKeyType
	}
	if len(v.data) == 0 {
		fmt.Fprintf(os.Stderr, "Error deleting item with %s=%s: Item not found.\n", keytype, key)
		return errNotFound
	}
	if len(v.data) == 1 {
		if (keytype == KeyMAC && v.data[0].MAC == key) ||
		   (keytype == KeyName && v.data[0].Name == key) ||
		   (keytype == KeyIP && v.data[0].IP == key) {
			v.data = nil
			return nil
		}
		fmt.Fprintf(os.Stderr, "Error deleting item with %s=%s: Item not found.\n", keytype, key)
		return errNotFound
	}
	found := -1
	for n, c := range(v.data) {
		if (keytype == KeyMAC && c.MAC == key) ||
		   (keytype == KeyName && c.Name == key) ||
		   (keytype == KeyIP && c.IP == key) {
			if found >= 0 {
				return errNotUnique
		fmt.Fprintf(os.Stderr, "Error deleting item with %s=%s: Multiple items found.\n", keytype, key)
			}
			found = n
		}
	}
	if found >= 0 {
		v.data[found] = v.data[len(v.data)-1]
		v.data = v.data[:len(v.data)-1]
		return nil
	}

	fmt.Fprintf(os.Stderr, "Error deleting item with %s=%s: Item not found.\n", keytype, key)
	return errNotFound
}

func (v *volatileStore) Assign (keytype, key, assignee string) error {
	if assignee != "" && len(assignee) != 3 {
		fmt.Fprintf(os.Stderr, "Error assigning item: Assignee code must be exactly 3 characters long.\n")
		return errMalformed
	}
	if keytype != KeyMAC && keytype != KeyName && keytype != KeyIP {
		if keytype == KeyAssignee || keytype == KeyNotAssigned {
			fmt.Fprintf(os.Stderr, "Error assigning item: Invalid key type %s.\n", keytype)
			return errInvalidKeyType
		}
		fmt.Fprintf(os.Stderr, "Error assigning item: Unknown key type %s.\n", keytype)
		return errUnknownKeyType
	}
	found := -1
	for n, _ := range(v.data) {
		if (keytype == KeyMAC && v.data[n].MAC == key) ||
		   (keytype == KeyName && v.data[n].Name == key) ||
		   (keytype == KeyIP && v.data[n].IP == key) {
			if found >= 0 {
				fmt.Fprintf(os.Stderr, "Error assigning item with %s=%s: Multiple items found.\n", keytype, key)
				return errNotUnique
			}
			found = n
		}
	}
	if found >= 0 {
		v.data[found].Assignee = assignee
		return nil
	}

	fmt.Fprintf(os.Stderr, "Error assigining item with %s=%s: Item not found.\n", keytype, key)
	return errNotFound
}

func (v *volatileStore) Unassign (keytype, key string) error {
	return v.Assign(keytype, key, "")
}

func (v *volatileStore) Close() error {
	return nil
}

/********/
/* JSON */
/********/

type jsonStore struct {
	file	*os.File
	v	dataInterface
}

var j jsonStore
func initJSON (jp *dataInterface, filename string) error {
	var err error

	err = initVolatile(&j.v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing internal database: %s\n", err.Error())
		return errCreatingDB
	}

	// Check if file exists
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		j.file, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating file %s: %s\n", filename, err.Error())
			return errCreatingDB
		}
	} else {

		j.file, err = os.OpenFile(filename, os.O_RDWR, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file %s: %s\n", filename, err.Error())
			return errOpeningDB
		}

		dec := json.NewDecoder(j.file)
		_, err = dec.Token()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding JSON: %s\n", err.Error())
			return errReadingDB
		}
		for dec.More() {
			var c Computer
			err := dec.Decode(&c)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error decoding JSON: %s\n", err.Error())
				return errReadingDB
			}
			err = j.v.Add(c)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error updating internal database: %s\n", err.Error())
				return errReadingDB
			}
		}
		_, err = dec.Token()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding JSON: %s\n", err.Error())
			return errReadingDB
		}
	}

	*jp = (dataInterface)(&j)

	return nil
}

func (j *jsonStore) Read (keytype, key string) (error, *Computer) {
	return j.v.Read(keytype, key)
}

func (j *jsonStore) ReadAll (keytype, key string) (error, []Computer) {
	return j.v.ReadAll(keytype, key)
}

func (j *jsonStore) Write () error {
	j.file.Truncate(0)
	j.file.Seek(0, 0)
	err, cl := j.v.ReadAll(KeyAll, "")
	if len(cl) > 0 {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading internal database: %s\n", err.Error())
			return err
		}
		err = json.NewEncoder(j.file).Encode(cl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %s\n", err.Error())
			return errWritingDB
		}
	} else {
		// Ignore error
		n, err := j.file.WriteString("[]")
		if n != 2 || err != nil {
			fmt.Fprintf(os.Stderr, "Error writing JSON file: %s\n", err.Error())
			return errWritingDB
		}
	}
	return nil
}

func (j *jsonStore) Add (c Computer) error {
	err := j.v.Add(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding item to internal database: %s\n", err.Error())
		return err
	}
	return j.Write()
}

func (j *jsonStore) Delete (keytype, key string) error {
	err := j.v.Delete(keytype, key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting item from internal database: %s\n", err.Error())
		return err
	}
	return j.Write()
}

func (j *jsonStore) Assign (keytype, key, assignee string) error {
	err := j.v.Assign(keytype, key, assignee)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating item assignment in internal database: %s\n", err.Error())
		return err
	}
	return j.Write()
}

func (j *jsonStore) Unassign (keytype, key string) error {
	err := j.v.Assign(keytype, key, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error removing item assignment in internal database: %s\n", err.Error())
		return err
	}
	return j.Write()
}

func (j *jsonStore) Close() error {
	err := j.file.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error closing internal database: %s\n", err.Error())
		return errClosingDB
	}
	return nil
}

/***********/
/* SQLite */
/***********/

type sqlStore struct {
	data *sql.DB
}

var db sqlStore

func initSQL (sqlp *dataInterface, filename string) error {
	var err error

	db.data, err = sql.Open("sqlite3", filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening SQL database: %s\n", err.Error())
		return errOpeningDB
	}
	db.data.SetMaxOpenConns(1)

	createSQL := `CREATE TABLE IF NOT EXISTS computers (
		MAC VARCHAR(17) NOT NULL,
		Name VARCHAR(50) NOT NULL,
  		IP VARCHAR(15) NOT NULL,
		Assignee VARCHAR(3),
		Description TEXT,
		PRIMARY KEY (MAC, Name, IP)
	);`

	_, err = db.data.Exec(createSQL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating table in SQL database: %s\n", err.Error())
		return errCreatingDB
	}

	*sqlp = (dataInterface)(&db)

	return nil
}

func resolveTransaction(tx *sql.Tx) {
	if p := recover(); p != nil {
		fmt.Fprintf(os.Stderr, "Error encountered while updating SQL database. Rolling it back.\n")
		tx.Rollback()
		return
	}
	tx.Commit()
}

func (db *sqlStore) Read (keytype, key string) (error, *Computer) {
	if keytype != KeyMAC && keytype != KeyName && keytype != KeyIP {
		if keytype == KeyAssignee || keytype == KeyNotAssigned {
			fmt.Fprintf(os.Stderr, "Error fetching item: Unknown key type %s.\n", keytype)
			return errInvalidKeyType, nil
		}
		fmt.Fprintf(os.Stderr, "Error fetching item: Unknown key type %s.\n", keytype)
		return errUnknownKeyType, nil
	}
	selectSQL := fmt.Sprintf("SELECT * FROM computers WHERE %s='%s'", keytype, key)
	rows, err := db.data.Query(selectSQL)
	defer rows.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching item: %s\n", err.Error())
		return errReadingDB, nil
	}

	var c Computer
	var assignee, description sql.NullString
	if rows.Next() {
		err = rows.Scan(&c.MAC, &c.Name, &c.IP, &assignee, &description)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching item: %s\n", err.Error())
			return errReadingDB, nil
		}
	} else {
		fmt.Fprintf(os.Stderr, "Error fetching item with %s=%s: Item not found.\n", keytype, key)
		return errNotFound, nil
	}
	if assignee.Valid {
		c.Assignee = assignee.String
	} else {
		c.Assignee = ""
	}
	if description.Valid {
		c.Description = description.String
	} else {
		c.Description = ""
	}
	if rows.Next() {
		fmt.Fprintf(os.Stderr, "Error fetching item with %s=%s: Mutlitple items found.\n", keytype, key)
		return errNotUnique, nil
	}
	return nil, &c
}

func (db *sqlStore) ReadAll (keytype, key string) (error, []Computer) {
	var selectSQL string

	if keytype == KeyAssignee {
		selectSQL = fmt.Sprintf("SELECT * FROM computers WHERE Assignee = '%s'", key)
	} else if keytype == KeyNotAssigned {
		selectSQL = fmt.Sprintf("SELECT * FROM computers WHERE Assignee = '' OR Assignee IS NULL")
	} else if keytype == KeyAll {
		selectSQL = fmt.Sprintf("SELECT * FROM computers")
	} else if keytype == KeyMAC || keytype == KeyName || keytype == KeyIP {
		fmt.Fprintf(os.Stderr, "Error fetching items: Invalid key type %s.\n", keytype)
		return errInvalidKeyType, nil
	} else {
		fmt.Fprintf(os.Stderr, "Error fetching items: Unknown key type %s.\n", keytype)
		return errUnknownKeyType, nil
	}

	rows, err := db.data.Query(selectSQL)
	defer rows.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading database: %s.\n", err.Error())
		return errReadingDB, nil
	}

	var cl []Computer
	for rows.Next() {
		var c Computer
		var assignee, description sql.NullString
		err = rows.Scan(&c.MAC, &c.Name, &c.IP, &assignee, &description)
		if err == nil {
			if assignee.Valid {
				c.Assignee = assignee.String
			} else {
				c.Assignee = ""
			}
			if description.Valid {
				c.Description = description.String
			} else {
				c.Description = ""
			}
			cl = append(cl, c)
		}
	}

	if len(cl) == 0 {
		fmt.Fprintf(os.Stderr, "Error fetching item with %s=%s: Item not found.\n", keytype, key)
		return errNotFound, nil
	}

	return nil, cl

}

func (db *sqlStore) Add (c Computer) error {
	if c.MAC == "" || c.Name == "" || c.IP == "" {
		fmt.Fprintf(os.Stderr, "Error adding item: MAC, Name and IP are mandatory fields.\n")
		return errMalformed
	}
	if c.Assignee != "" && len(c.Assignee) != 3 {
		fmt.Fprintf(os.Stderr, "Error adding item: Assignee code must be exactly 3 characters long.\n")
		return errMalformed
	}
	tx, err := db.data.Begin()
	defer resolveTransaction(tx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to database: %s\n", err.Error())
		return errWritingDB
	}
	stmt, err := tx.Prepare("INSERT INTO computers(MAC, Name, IP, Assignee, Description) VALUES (?, ?, ?, ?, ?)")
	defer stmt.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to database: %s\n", err.Error())
		return errWritingDB
	}
	_, err = stmt.Exec(c.MAC, c.Name, c.IP, c.Assignee, c.Description)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to database: %s\n", err.Error())
		return errWritingDB
	}

	return nil
}

func (db *sqlStore) Delete (keytype, key string) error {
	if keytype != KeyMAC && keytype != KeyName && keytype != KeyIP {
		if keytype == KeyAssignee || keytype == KeyNotAssigned {
			fmt.Fprintf(os.Stderr, "Error deleting item: Invalid key type %s.\n", keytype)
			return errInvalidKeyType
		}
		fmt.Fprintf(os.Stderr, "Error deleting item: Unknown key type %s.\n", keytype)
		return errUnknownKeyType
	}

	deleteSQL := fmt.Sprintf("DELETE FROM computers WHERE %s = ?", keytype)
	tx, err := db.data.Begin()
	defer resolveTransaction(tx)
        if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to database: %s\n", err.Error())
                return errWritingDB
        }

	stmt, err := tx.Prepare(deleteSQL)
        defer stmt.Close()
        if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to database: %s\n", err.Error())
                return errWritingDB
        }

        _, err = stmt.Exec(key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to database: %s\n", err.Error())
		return errWritingDB
	}

	return nil

}

func (db *sqlStore) Assign (keytype, key, assignee string) error {
	if keytype != KeyMAC && keytype != KeyName && keytype != KeyIP {
		if keytype == KeyAssignee || keytype == KeyNotAssigned {
			fmt.Fprintf(os.Stderr, "Error assigning item: Invalid key type %s.\n", keytype)
			return errInvalidKeyType
		}
		fmt.Fprintf(os.Stderr, "Error assigning item: Unknown key type %s.\n", keytype)
		return errUnknownKeyType
	}
	if len(assignee) != 3 {
		fmt.Fprintf(os.Stderr, "Error assigning item: Assignee code must be exactly 3 characters long.\n")
		return errMalformed
	}
	tx, err := db.data.Begin()
	defer resolveTransaction(tx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to database: %s\n", err.Error())
		return errWritingDB
	}
	updateSQL := fmt.Sprintf("UPDATE computers SET Assignee = ? WHERE %s = ?", keytype)
	stmt, err := tx.Prepare(updateSQL)
	defer stmt.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to database: %s\n", err.Error())
		return errWritingDB
	}
	_, err = stmt.Exec(assignee, key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to database: %s\n", err.Error())
		return errWritingDB
	}

	return nil
}

func (db *sqlStore) Unassign (keytype, key string) error {
	if keytype != KeyMAC && keytype != KeyName && keytype != KeyIP {
		if keytype == KeyAssignee || keytype == KeyNotAssigned {
			fmt.Fprintf(os.Stderr, "Error removing assignment: Invalid key type %s.\n", keytype)
			return errInvalidKeyType
		}
		fmt.Fprintf(os.Stderr, "Error removing assignment: Unknown key type %s.\n", keytype)
		return errUnknownKeyType
	}
	tx, err := db.data.Begin()
	defer resolveTransaction(tx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to database: %s\n", err.Error())
		return errWritingDB
	}
	updateSQL := fmt.Sprintf("UPDATE computers SET Assignee = '' WHERE %s = ?", keytype)
	stmt, err := tx.Prepare(updateSQL)
	defer stmt.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to database: %s\n", err.Error())
		return errWritingDB
	}
	_, err = stmt.Exec(key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to database: %s\n", err.Error())
		return errWritingDB
	}
	return nil
}

func (db *sqlStore) Close() error {
	err := db.data.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error closing database: %s\n", err.Error())
		return errClosingDB
	}
	return nil
}
