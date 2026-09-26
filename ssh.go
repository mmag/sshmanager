package main

import (
	"context"
	"errors"
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

// sshCheck reports whether sshd answers, trying to log in with ssh in batch mode.
// Unlike ping it follows ~/.ssh/config (aliases, ProxyJump) and works for hosts
// that block ICMP. A rejected login still means the server is alive.
func sshCheck(conn SSHConnection) hostStatus {
	ctx, cancel := context.WithTimeout(context.Background(), sshCheckTimeout)
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

	output, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		return statusOffline
	}

	// ssh exits with 255 on its own errors, any other code comes from the remote command
	var exitErr *exec.ExitError
	if err == nil || errors.As(err, &exitErr) && exitErr.ExitCode() != 255 || sshdResponded(string(output)) {
		return statusOnline
	}
	return statusOffline
}

// sshdResponded reports whether ssh failed after reaching the server
func sshdResponded(output string) bool {
	for _, marker := range []string{
		"Permission denied",
		"Host key verification failed",
		"Too many authentication failures",
		"Unable to negotiate",
	} {
		if strings.Contains(output, marker) {
			return true
		}
	}
	return false
}
