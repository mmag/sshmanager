package main

import (
	"context"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	pingTimeout = 2 * time.Second
	// minCheckInterval protects hosts from a too small check_interval in the config
	minCheckInterval = 5 * time.Second
)

type hostStatus int

const (
	statusUnknown hostStatus = iota
	statusOnline
	statusOffline
)

var (
	// hostStatuses and the in-flight sets are keyed by server address
	// and only accessed from the UI goroutine
	hostStatuses   = make(map[string]hostStatus)
	checksInFlight = make(map[string]bool)
	// loginChecksInFlight holds hosts checked with ssh; pings don't touch their status meanwhile
	loginChecksInFlight = make(map[string]bool)

	checkTicker   *time.Ticker
	checkInterval time.Duration
)

func statusSymbol(server string) string {
	switch hostStatuses[server] {
	case statusOnline:
		return "[green]✓[-]"
	case statusOffline:
		return "[red]✗[-]"
	}
	return "[gray]?[-]"
}

// connectionHost extracts the host name or IP from the server field,
// which may contain a user@ prefix, a :port suffix or IPv6 brackets
func connectionHost(conn SSHConnection) string {
	host := strings.TrimSpace(conn.Server)
	if at := strings.LastIndex(host, "@"); at >= 0 {
		host = host[at+1:]
	}
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		return parsedHost
	}
	return strings.TrimSuffix(strings.TrimPrefix(host, "["), "]")
}

// pingCommand builds a single ping with a timeout for the current platform
func pingCommand(ctx context.Context, host string) *exec.Cmd {
	seconds := strconv.Itoa(int(pingTimeout.Seconds()))
	switch runtime.GOOS {
	case "windows":
		return exec.CommandContext(ctx, "ping", "-n", "1", "-w", strconv.Itoa(int(pingTimeout.Milliseconds())), host)
	case "darwin", "freebsd":
		// ping there doesn't accept IPv6 addresses, and ping6 has no timeout option
		if strings.Contains(host, ":") {
			return exec.CommandContext(ctx, "ping6", "-c", "1", host)
		}
		return exec.CommandContext(ctx, "ping", "-c", "1", "-t", seconds, host)
	default:
		return exec.CommandContext(ctx, "ping", "-c", "1", "-W", seconds, host)
	}
}

// pingHost reports whether the host answers an ICMP echo request.
// Ping doesn't touch sshd, so frequent checks don't trigger fail2ban and similar tools.
func pingHost(host string) bool {
	if host == "" {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout+time.Second)
	defer cancel()
	return pingCommand(ctx, host).Run() == nil
}

// checkHosts pings the given connections in the background and updates
// their statuses as results arrive. Hosts still being checked are skipped.
func checkHosts(connections []SSHConnection) {
	for _, conn := range connections {
		server := conn.Server
		if checksInFlight[server] || loginChecksInFlight[server] {
			continue
		}
		checksInFlight[server] = true

		host := connectionHost(conn)
		go func() {
			status := statusOffline
			if pingHost(host) {
				status = statusOnline
			}
			app.QueueUpdateDraw(func() {
				delete(checksInFlight, server)
				if !connectionExists(server) || loginChecksInFlight[server] || hostStatuses[server] == status {
					return
				}
				hostStatuses[server] = status
				updateStatusSymbols()
			})
		}()
	}
}

// checkLogin checks the host with a full ssh login in the background;
// the status shows as unknown until the result arrives
func checkLogin(conn SSHConnection) {
	server := conn.Server
	if loginChecksInFlight[server] {
		return
	}
	loginChecksInFlight[server] = true
	hostStatuses[server] = statusUnknown
	updateStatusSymbols()

	go func() {
		status := sshCheck(conn)
		app.QueueUpdateDraw(func() {
			delete(loginChecksInFlight, server)
			if connectionExists(server) {
				hostStatuses[server] = status
				updateStatusSymbols()
			}
		})
	}()
}

// startPeriodicChecks rechecks all hosts every check_interval seconds; 0 disables it
func startPeriodicChecks() {
	if config.CheckInterval <= 0 {
		return
	}
	checkInterval = max(time.Duration(config.CheckInterval)*time.Second, minCheckInterval)
	checkTicker = time.NewTicker(checkInterval)
	go func() {
		for range checkTicker.C {
			app.QueueUpdate(func() {
				checkHosts(allConnections())
			})
		}
	}()
}

// pauseChecks stops periodic checks while an ssh session occupies the terminal.
// Otherwise updates would pile up in the event queue, which is not processed while suspended.
func pauseChecks() {
	if checkTicker != nil {
		checkTicker.Stop()
	}
}

// resumeChecks restarts periodic checks and refreshes all statuses right away
func resumeChecks() {
	if checkTicker != nil {
		checkTicker.Reset(checkInterval)
	}
	checkHosts(allConnections())
}
