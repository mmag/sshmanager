package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	sshCheckTimeout   = 15 * time.Second
	sshConnectTimeout = "5"
)

// sshArgs returns the port option and user@host target for the ssh command
func sshArgs(conn SSHConnection) []string {
	var args []string
	if conn.Port != "" {
		args = append(args, "-p", conn.Port)
	}
	target := conn.Server
	if conn.Username != "" {
		target = conn.Username + "@" + conn.Server
	}
	return append(args, target)
}

// sshConnect runs an interactive ssh session in the current terminal
func sshConnect(conn SSHConnection) {
	cmd := exec.Command("ssh", sshArgs(conn)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf(currentLang["msg_connecting"], conn.Server)
	if err := cmd.Run(); err != nil {
		log.Printf(currentLang["msg_conn_error"], conn.Server, err)
	}
}

// sshCheck logs in with ssh in batch mode and describes the outcome.
// Unlike ping it follows ~/.ssh/config (aliases, ProxyJump) and verifies
// that sshd accepts the user's key.
func sshCheck(ctx context.Context, conn SSHConnection) string {
	ctx, cancel := context.WithTimeout(ctx, sshCheckTimeout)
	defer cancel()

	args := []string{
		"-n", "-T",
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=" + sshConnectTimeout,
		"-o", "ControlPath=none",
		"-o", "LogLevel=ERROR",
	}
	args = append(args, sshArgs(conn)...)
	cmd := exec.CommandContext(ctx, "ssh", append(args, "exit")...)
	cmd.WaitDelay = time.Second

	start := time.Now()
	output, err := cmd.CombinedOutput()
	elapsed := time.Since(start)

	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Sprintf(currentLang["ssh_timeout"], conn.Server, int(sshCheckTimeout.Seconds()))
	}

	// ssh exits with 255 on its own errors, any other code comes from the remote command
	var exitErr *exec.ExitError
	if err == nil || errors.As(err, &exitErr) && exitErr.ExitCode() != 255 {
		return fmt.Sprintf(currentLang["ssh_ok"], conn.Server, elapsed.Seconds())
	}

	message := lastLine(string(output))
	if message == "" {
		message = err.Error()
	}
	if sshdResponded(message) {
		return fmt.Sprintf(currentLang["ssh_denied"], conn.Server, message)
	}
	return fmt.Sprintf(currentLang["ssh_failed"], conn.Server, message)
}

// sshdResponded reports whether an ssh error happened after reaching the server
func sshdResponded(message string) bool {
	for _, marker := range []string{
		"Permission denied",
		"Host key verification failed",
		"Too many authentication failures",
		"Unable to negotiate",
	} {
		if strings.Contains(message, marker) {
			return true
		}
	}
	return false
}

func lastLine(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}
