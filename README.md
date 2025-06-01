# vai-vacuum

A tool to create VACUUM snapshots of the VAI (Virtual Aggregate Informer) database from Rancher pods.

## What it does

- Opens the VAI database at `/var/lib/rancher/informer_object_cache.db` (read-only)
- Creates a VACUUM snapshot (compact copy) of the database
- Outputs the entire database as base64 to stdout
- All errors go to stderr prefixed with "ERROR:"

## Building

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o vai-vacuum .
```

This creates a static Linux binary (~6MB) that runs in Rancher pods without any dependencies.

## Deployment

Do NOT commit the binary to git. Instead:

1. Build the binary using the command above
2. Create a GitHub Release in your repository
3. Upload `vai-vacuum` as a release asset
4. Use the release download URL in your tests

Example URL format:
```
https://github.com/YOUR-ORG/YOUR-REPO/releases/download/v1.0.0/vai-vacuum
```

## Usage

The binary is used by VAI tests to extract database snapshots from running Rancher pods:

```bash
# Download and run on a pod
kubectl exec <pod> -n cattle-system -c rancher -- sh -c \
  "curl -L -o /tmp/vai-vacuum <RELEASE-URL> && chmod +x /tmp/vai-vacuum && /tmp/vai-vacuum"
```

The output is pure base64 which can be decoded to get the SQLite database file.

## Dependencies

- `modernc.org/sqlite` - Pure Go SQLite driver (no CGO required)

## Output Format

- **Success**: Raw base64 encoded database to stdout (no headers, no newlines at start)
- **Failure**: Error message to stderr starting with "ERROR:"

## Local Testing

To test on your development machine:

### 1. Modify the source code for local testing

⚠️ **Important**: The code uses a production path that doesn't exist locally. You need to temporarily modify it.

In your Go source file, change:
```go
const (
    vaiDBPath = "/var/lib/rancher/informer_object_cache.db"
```

To:
```go
const (
    vaiDBPath = "./test.db"  // Temporary change for local testing
```

### 2. Build for your platform

```bash
# For macOS (Intel)
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o vai-vacuum-mac .

# For macOS (Apple Silicon M1/M2)
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o vai-vacuum-mac .
```

### 3. Test locally

```bash
# Create test database
sqlite3 test.db "CREATE TABLE test (id INTEGER); INSERT INTO test VALUES (1);"

# Run the tool
./vai-vacuum-mac | base64 -d > output.db

# Verify the output
sqlite3 output.db "SELECT * FROM test;"
# Should output: 1
```

### 4. Revert code and build for production

⚠️ **Don't forget**: Change the source code back to the production path:

```go
const (
    vaiDBPath = "/var/lib/rancher/informer_object_cache.db"
```

Then build the Linux version for deployment:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o vai-vacuum .
```

This creates the Linux binary that will run in Rancher pods.

## Notes

- The tool uses SQLite's VACUUM command to create a compact copy of the database
- The snapshot is created in `/tmp/vai-snapshot.db` temporarily and cleaned up after encoding
- All database operations are read-only to ensure safety in production