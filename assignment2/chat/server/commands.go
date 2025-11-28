package server

import (
        "bufio"
        "fmt"
        "net"
        "strings"
)

// commands.go implements all chat commands (/nck, /lst, /msg, /grp)

// handleCommand parses a line for a command and executes it
func handleCommand(conn net.Conn, w *bufio.Writer, state *ChatState, line string) {
        line = strings.TrimSpace(line)
        if line == "" {
                return
        }

        // split into command and args
        var cmd, args string

        if space := strings.IndexByte(line, ' '); space == -1 {
                cmd = line
                args = ""
        } else {
                cmd = line[:space]
                args = strings.TrimSpace(line[space+1:])
        }

        cmd_lower := strings.ToLower(cmd)

        switch cmd_lower {
        case "/nck":
                setNickname(conn, w, state, args)
        case "/lst":
                listUsers(conn, w, state)
        case "/msg":
                sendMessage(conn, w, state, args)
        case "/grp":
                createGroup(conn, w, state, args)
        default:
                w.WriteString("Unknown command\n")
		w.Flush()
        }
}

// registers a nickname for a connection
func setNickname(conn net.Conn, writer *bufio.Writer, state *ChatState, nickname string) {
        state.mtx.Lock()
        defer state.mtx.Unlock()

        if nickname == "" {
                writer.WriteString("Nickname value cannot be empty\n")
                writer.Flush()
                return
        }
        if len(nickname) > 10 {
                writer.WriteString("Nickname too long\n")
                writer.Flush()
                return
        }
        if _, exists := state.nicknames[nickname]; exists {
                writer.WriteString("Nickname already taken\n")
                writer.Flush()
                return
        }

	// user rename then replace old
        if old, ok := state.owners[conn]; ok {
                delete(state.nicknames, old)
        }

	// save nickname so others can message this connection, conn will own this nickname
        state.nicknames[nickname] = conn
        state.owners[conn] = nickname

        // ensure groups map is initialized
        if _, ok := state.groups[conn]; !ok {
                state.groups[conn] = make(map[string][]string)
        }

        writer.WriteString(fmt.Sprintf("Nickname set to %s\n", nickname))
        writer.Flush()
}

// returns connected users
func listUsers(conn net.Conn, writer *bufio.Writer, state *ChatState) {
        state.mtx.Lock()
        defer state.mtx.Unlock()

        user_list := make([]string, 0, len(state.nicknames))
        for n := range state.nicknames {
                user_list = append(user_list, n)
        }

        writer.WriteString(strings.Join(user_list, ", ") + "\n")
        writer.Flush()
}

// delivers a message
func sendMessage(conn net.Conn, writer *bufio.Writer, state *ChatState, arg string) {
	state.mtx.Lock()
	defer state.mtx.Unlock()

	// sender must have a nickname
	sender, ok := state.owners[conn]
	if !ok {
		writer.WriteString("Set nickname first\n")
		writer.Flush()
		return
	}

	// split <recipients> and <msg>
	parts := strings.SplitN(arg, " ", 2)
	if len(parts) != 2 {
		writer.WriteString("Usage: /MSG <to> <msg>\n")
		writer.Flush()
		return
	}

	recip_str := parts[0]
	msg_text := parts[1]

	// expand comma separated recipients/groups
	raw := strings.Split(recip_str, ",")
	recipients := make([]string, 0)
	used_group := ""

	// expand groups (#groupname)
	for _, r := range raw {
		r = strings.TrimSpace(r)

		// groups r has # 
		if strings.HasPrefix(r, "#") {
			if group_members, exists := state.groups[conn][r]; exists {
				used_group = r
				recipients = append(recipients, group_members...)
			}
		} else {
			recipients = append(recipients, r)
		}
	}

	delivered := 0
	// make sure msg is only sent once to stop like /msg atty,atty 
	unique_seen := make(map[string]bool)

	// deliver message to each resolved recipient
	for _, r := range recipients {
		if r == "" || r == sender {
			continue
		}
		if !unique_seen[r] {
			unique_seen[r] = true
			if to_conn, exists := state.nicknames[r]; exists {
				if used_group == "" {
					// direct message
					fmt.Fprintf(to_conn, "%s: %s\n", sender, msg_text)
				} else {
					// group formatted delivery
					fmt.Fprintf(to_conn, "[%s] %s: %s\n", used_group[1:], sender, msg_text)
				}
				delivered++
			}
		}
	}

	if delivered == 0 {
		writer.WriteString("No valid recipients\n")
	} else {
		writer.WriteString("Message sent\n")
	}

	// echo /msg #grp back to sender
	if used_group != "" {
		echo := fmt.Sprintf("[%s] %s: %s\n", used_group[1:], sender, msg_text)
		writer.WriteString(echo)
	}

	writer.Flush()
}


// registers a new group
func createGroup(conn net.Conn, writer *bufio.Writer, state *ChatState, arg string) {
        parts := strings.SplitN(arg, " ", 2)
        if len(parts) != 2 {
                writer.WriteString("Usage: /GRP <#group> <user1,user2,...>\n")
                writer.Flush()
                return
        }

        group := parts[0]
        members := strings.Split(parts[1], ",")

        if !strings.HasPrefix(group, "#") || len(group) > 11 {
                writer.WriteString("Invalid group name\n")
                writer.Flush()
                return
        }

        state.mtx.Lock()
        defer state.mtx.Unlock()

        if _, ok := state.groups[conn]; !ok {
                state.groups[conn] = make(map[string][]string)
        }

        cleaned := []string{}
        for _, m := range members {
                nick := strings.TrimSpace(m)
                if nick != "" {
                        cleaned = append(cleaned, nick)
                }
        }

        state.groups[conn][group] = cleaned

        writer.WriteString(fmt.Sprintf("Group %s created\n", group))
        writer.Flush()
}

