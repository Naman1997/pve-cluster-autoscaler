package main

import (
	"crypto/tls"
	"database/sql"
	"log"
	"os"
	"strconv"

	"github.com/Telmate/proxmox-api-go/proxmox"
)

/*
validateInputs validates that all
required inputs are in place and
 are using the correct formats.
*/
func validateInputs() (int, *tls.Config, string, string, int, int) {
	insecure, err := strconv.ParseBool(getValueOf("insecure", "false"))
	FailError(err)
	*proxmox.Debug, err = strconv.ParseBool(getValueOf("debug", "false"))
	FailError(err)
	taskTimeout, err := strconv.Atoi(getValueOf("taskTimeout", "300"))
	FailError(err)
	memLimit := getValueOf("memoryLimit", "")
	if len(memLimit) == 0 {
		log.Fatal("memoryLimit not specified in config!")
	}
	memoryLimit, err := strconv.Atoi(memLimit)
	FailError(err)
	cLimit := getValueOf("cpuLimit", "")
	if len(cLimit) == 0 {
		log.Fatal("cpuLimit not specified in config!")
	}
	cpuLimit, err := strconv.Atoi(cLimit)
	FailError(err)
	node := getValueOf("nodeName", "")
	if len(node) == 0 {
		log.Fatal("Node name not specified in config!")
	}
	template := getValueOf("templateName", "")
	if len(template) == 0 {
		log.Fatal("Template name not specified in config!")
	}
	tlsconf := &tls.Config{InsecureSkipVerify: true}
	if !insecure {
		tlsconf = nil
	}
	return taskTimeout, tlsconf, template, node, cpuLimit, memoryLimit
}

/*
validatePostgresConfig validates that
vars provided for postgres connection
are present and tries to test the
connection to postgres DB
*/
func validatePostgresConfig() string {
	dbName := os.Getenv("POSTGRES_DB")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")

	if len(dbName) == 0 {
		log.Fatal("Database name not specified!")
	}
	if len(user) == 0 {
		log.Fatal("Database user not specified!")
	}
	if len(password) == 0 {
		log.Fatal("Database password not specified!")
	}
	return testDBConnection(dbName, user, password)
}

/*
testDBConnection pings the db and
makes sure the creds are valid by
creating the vms table
*/
func testDBConnection(dbName string, user string, password string) string {
	connStr := "host=postgres-db-lb port=5432 user=" + user + " dbname=" + dbName + " sslmode=disable password=" + password
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		FailError(err)
	}
	err = db.Ping()
	if err != nil {
		FailError(err)
	}
	createTable(db)
	defer db.Close()
	return connStr
}

// Creates the vms table in postgres
func createTable(db *sql.DB) error {
	sqlStatement := `CREATE TABLE IF NOT EXISTS vms (vmid serial PRIMARY KEY,
					node VARCHAR(50) NOT NULL,
					pool VARCHAR(50),
					vmType VARCHAR(50),
					memory INTEGER NOT NULL,
					cores INTEGER NOT NULL
					);`

	_, err := db.Exec(sqlStatement)
	return err
}

func insertVmInfo(db *sql.DB, vmr *proxmox.VmRef, config *proxmox.ConfigQemu) error {
	sqlStatement := `INSERT INTO vms (vmid, node, pool, vmtype, memory, cores) VALUES ($1, $2, $3, $4, $5, $6) RETURNING vmid;`
	_, err := db.Exec(sqlStatement, vmr.VmId(), vmr.Node(), config.Pool, vmr.GetVmType(), config.Memory, config.QemuCores)
	if err != nil {
		ColorPrint(INFO, "Ran into error while insering data into db: %v", err)
		ColorPrint(WARN, "Attempting to re-create vms table if it does not exists.")
		dbErr := createTable(db)
		if dbErr != nil {
			ColorPrint(INFO, "Ran into error while re-creating vms table: %v", dbErr)
		} else {
			ColorPrint(WARN, "'vms' table was re-created in DB!")
			ColorPrint(WARN, "Application will not have access to older VM records if the table was deleted manually!")
		}
	} else {
		ColorPrint(INFO, "Saved config of cloned VM in DB. Id: %d", vmr.VmId())
		ColorPrint(INFO, "Info Saved: %d, %s", vmr.VmId(), vmr.Node())
	}
	return err
}
