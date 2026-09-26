package main

import (
	"net"
	"strings"
	"time"
)

const hostTimeout = 2 * time.Second

// hostOnline holds the last check result per server address.
// It is only accessed from the UI goroutine.
var hostOnline = make(map[string]bool)

func statusSymbol(server string) string {
	if hostOnline[server] {
		return "[green]✓[-]"
	}
	return "[red]✗[-]"
}

func connectionAddress(conn SSHConnection) string {
	host := strings.TrimSpace(conn.Server)
	if at := strings.LastIndex(host, "@"); at >= 0 {
		host = host[at+1:]
	}
	if host == "" {
		return ""
	}

	if conn.Port == "" {
		if parsedHost, parsedPort, err := net.SplitHostPort(host); err == nil && parsedHost != "" && parsedPort != "" {
			return net.JoinHostPort(parsedHost, parsedPort)
		}
		if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
			host = strings.TrimPrefix(strings.TrimSuffix(host, "]"), "[")
		}
		return net.JoinHostPort(host, "22")
	}

	return net.JoinHostPort(host, conn.Port)
}

func checkHostOnline(conn SSHConnection) bool {
	address := connectionAddress(conn)
	if address == "" {
		return false
	}

	connection, err := net.DialTimeout("tcp", address, hostTimeout)
	if err != nil {
		return false
	}
	_ = connection.Close()
	return true
}

// checkHosts probes the given connections in the background
// and refreshes the list as results arrive
func checkHosts(connections []SSHConnection) {
	for _, conn := range connections {
		go func() {
			online := checkHostOnline(conn)
			app.QueueUpdateDraw(func() {
				hostOnline[conn.Server] = online
				refreshConnectionsList(connectionsList.GetCurrentItem())
			})
		}()
	}
}
