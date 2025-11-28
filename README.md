# Go Concurrent Chat Server Overview

## Repository Structure
- `chat/app/server/main.go`: Minimal entry point that boots the TCP chat server on port 6666 by calling `server.Start`.
- `chat/app/client/main.go`: CLI chat client that connects to the server at `127.0.0.1:6666`, starts a background listener for server output, and forwards user input lines to the server.
- `chat/server/`: Core server implementation shared by the app entry point.
  - `server.go`: Listens for TCP connections, initializes shared state, and spawns `HandleClient` goroutines.
  - `state.go`: Defines the `ChatState` structure and constructor for nicknames, ownership mapping, and per-connection groups with mutex protection.
  - `client_handler.go`: Manages per-client I/O loop and ensures cleanup on disconnect.
  - `commands.go`: Parses and executes chat commands (`/nck`, `/lst`, `/msg`, `/grp`).

## Server Lifecycle
1. `Start(port int)` builds a TCP listener and logs the bound address.
2. A shared `ChatState` is created to coordinate nicknames and groups across connections.
3. The accept loop spawns `HandleClient` for each connection so clients run concurrently.

### ChatState Contents
- `nicknames`: map of nickname → connection for routing messages.
- `owners`: map of connection → nickname for validating senders and cleanup.
- `groups`: per-connection map of group name → member nicknames. Groups are private to the connection that created them.
- All maps are guarded by a mutex in every command handler to keep access thread-safe.

### Client Handling Flow
1. `HandleClient` sets up buffered reader/writer pairs and defers a `disconnect` cleanup.
2. It scans incoming lines from the connection and forwards each line to `handleCommand`.
3. `disconnect` removes the client’s nickname, ownership entry, and groups before closing the connection.

## Commands
Commands are case-insensitive and follow IRC-style semantics. Every handler locks the shared state while mutating or reading shared maps.

### `/nck <nickname>`
- Validates presence, length (≤10 chars), and uniqueness of the nickname.
- Removes any prior nickname for the connection, then registers the new one.
- Initializes the caller’s group map if missing.
- Response: `Nickname set to <nickname>`.

### `/lst`
- Compiles the current nickname list and returns a comma-separated string.
- Uses the length of `state.nicknames` to size the slice efficiently.

### `/msg <to>[,<to>...] <message>`
- Requires the sender to have a nickname; otherwise replies `Set nickname first`.
- Splits the first argument into comma-separated recipients and expands any `#group` belonging to the sender into its member list.
- Deduplicates recipients to avoid double-delivery and skips self-targeting.
- Sends either `sender: message` (direct) or `[group] sender: message` (group) to each resolved connection.
- Replies `Message sent` when at least one recipient is found, otherwise `No valid recipients`.
- Echoes group-formatted messages back to the sender when a group target was used.

### `/grp <#group> <user1,user2,...>`
- Validates the name starts with `#` and is at most 11 characters.
- Cleans the member list by trimming whitespace and dropping empty entries.
- Stores the group only under the creator’s connection, keeping groups private to that user.
- Response: `Group <#group> created`.

## Client Behavior
- Connects to the server on startup and prints “Connected to chat”.
- Starts a goroutine that continuously prints any line received from the server.
- Main loop prompts with `> `, reads user input, and forwards each line to the server with `fmt.Fprintln`.
- Exits when `os.Stdin` closes.

## Running the System
1. From `chat/app/server`, run `go run .` to start the server (default port 6666).
2. From `chat/app/client`, run `go run .` to start a client. Open multiple terminals for multiple clients.
3. Use `/nck` to set your nickname before sending `/msg` or creating groups.

## Key Concurrency Considerations
- Each client connection is handled in its own goroutine for concurrent chat sessions.
- Shared state is guarded by a mutex on every access to prevent data races.
- The accept loop is infinite; stop the server with Ctrl+C or terminate the process.
