package main

import (
	"log"
	"os"
	"os/exec"
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
