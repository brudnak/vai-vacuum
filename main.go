package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	_ "modernc.org/sqlite" // Pure Go SQLite driver - no CGO required
)

const (
	// Path to VAI database in Rancher pods
	vaiDBPath = "/var/lib/rancher/informer_object_cache.db"

	// Temporary path for VACUUM snapshot
	snapshotPath = "/tmp/vai-snapshot.db"
)

func main() {
	// Open the VAI database in read-only mode
	db, err := sql.Open("sqlite", vaiDBPath+"?mode=ro")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to open VAI database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Cannot connect to VAI database: %v\n", err)
		os.Exit(1)
	}

	// Create a VACUUM snapshot
	// VACUUM creates a compact copy of the database, which is perfect for our use case
	_, err = db.Exec(fmt.Sprintf("VACUUM INTO '%s'", snapshotPath))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create VACUUM snapshot: %v\n", err)
		os.Exit(1)
	}

	// Open the snapshot file
	snapshotFile, err := os.Open(snapshotPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to open snapshot file: %v\n", err)
		os.Exit(1)
	}
	defer snapshotFile.Close()

	// Clean up snapshot file when done
	defer os.Remove(snapshotPath)

	// Create base64 encoder writing to stdout
	encoder := base64.NewEncoder(base64.StdEncoding, os.Stdout)

	// Copy snapshot to encoder (which outputs to stdout)
	_, err = io.Copy(encoder, snapshotFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to encode snapshot: %v\n", err)
		os.Exit(1)
	}

	// Close encoder to flush any remaining bytes
	encoder.Close()

	// Success - no output except the base64 data already written
}
