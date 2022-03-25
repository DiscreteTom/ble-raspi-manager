package main

import (
	"bytes"
	"os/exec"
)

func runCommand(command string) (stdout string, err error) {
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if err := cmd.Run(); err != nil {
		return "", err
	}

	stdout = stdoutBuf.String()
	return stdout, nil
}
