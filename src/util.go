package main

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func FailError(err error) {
	if err != nil {
		ColorPrint(ERROR, err.Error())
	}
}

func getValueOf(key, fallback string) string {
	value, err := os.ReadFile("/etc/secrets/" + key)
	if err != nil {
		return fallback
	}
	data := string(value)
	data = strings.Trim(data, "\"")
	return data
}

var rxUserRequiresToken = regexp.MustCompile("[a-z0-9]+@[a-z0-9]+![a-z0-9]+")

func userRequiresAPIToken(userID string) bool {
	return rxUserRequiresToken.MatchString(userID)
}

func sendCommands(user string, addr string, command string) error {
	cmd0 := "ssh"
	cmd1 := user + "@" + addr
	cmd2 := "-f"

	commandArr := []string{cmd0, cmd1, cmd2}
	cmd := exec.Command(commandArr[0], commandArr...)
	ColorPrint(INFO, "Executing: "+strings.Join(commandArr, " "))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
