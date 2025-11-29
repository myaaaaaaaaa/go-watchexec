# go-watchexec

`go-watchexec` is a simple command-line tool that watches a directory for file changes and executes a command in response.

```bash
go install github.com/myaaaaaaaaa/go-watchexec/watchexec@latest
```

## Features

- **Command Execution:** Executes a given command when a file is changed, clearing the screen and showing a timestamp on each run.
- **Unix-style Modularity:** Streams modified filenames (line-by-line) to `stdout` when the output is piped or redirected. This allows `go-watchexec` to be used as a file change event source in a command-line pipeline.

## Usage

`go-watchexec` has two main modes of operation.

### 1. Command Execution Mode

Pass a command as an argument to have it automatically re-run on any file change.

**Syntax:**
```bash
watchexec [command]
```

**Examples:**
```bash
# Re-run Go tests on any file change
watchexec go test ./...
```
```bash
# Rebuild and run a server
watchexec "go build -o myapp && ./myapp"
```

### 2. Streaming Mode

If you pipe or redirect the output, `go-watchexec` enters streaming mode. Instead of running a command, it prints the path of each modified file to `stdout` on a new line. This allows you to compose `go-watchexec` with other command-line tools for more complex workflows.

**Example:**

A common use case is to pipe the stream of changed files to `xargs` to perform an action on each file. The following example reports the line count of only the Go files that have changed:
```bash
# Get the line count of each changed Go file
watchexec | grep '\.go$' | xargs -n1 wc -l
```
In this pipeline:
1.  `watchexec` prints every modified file path to standard output.
2.  `grep '\.go$'` filters this stream to only include files ending in `.go`.
3.  `xargs -n1 wc -l` receives the filtered paths and runs `wc -l` on each one individually.

## Implementation: A Polling-based Watcher

This tool uses a polling-based mechanism to watch for file changes, as opposed to relying on OS-specific filesystem notification events (like Linux's `inotify`). Here's how it works and why this approach was chosen:

1.  The watcher periodically scans the directory tree to get a list of all files.
2.  It then iterates through the files, checking the "last modified" timestamp of each one.
3.  If a timestamp is newer than the last one seen, it triggers the specified command.

**Benefits of this approach:**

*   **Portability:** It works on any operating system that Go supports without any extra dependencies. It does not rely on OS-specific APIs, making it universally compatible.
*   **Filesystem Support:** It functions reliably across a wide variety of filesystems, including network filesystems (NFS, Samba), where event-based watchers can be unreliable.
*   **Simplicity:** The polling approach provides robust and straightforward recursive directory monitoring. It automatically handles newly created directories without the complexity of adding new filesystem-level watches.

This design choice prioritizes portability and simplicity, ensuring that `go-watchexec` works consistently everywhere.
