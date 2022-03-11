package main

import (
	"os"
	"os/exec"
	"strings"
)

/*
AnsibleGalaxy executes ansible-galaxy collection install
using the requirements filepath passed
Expects ansible binary to be present
in PATH
*/
func AnsibleGalaxy(requirements string) {
	if len(requirements) > 0 {
		cmd0 := "ansible-galaxy"
		cmd1 := "collection"
		cmd2 := "install"
		cmd3 := "-r"
		cmd4 := strings.Trim(requirements, "\n")
		cmd := exec.Command(cmd0, cmd1, cmd2, cmd3, cmd4)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			ColorPrint(ERROR, "%v", err)
		}

		ColorPrint(INFO, "Finished installing ansible requirements")
	} else {
		ColorPrint(WARN, "Requirements file was not proivded! Skipping requirements installation for ansible")
	}
}

/*
AnsiblePlaybook executes ansible-playbook
using the playbook filepath passed
Expects ansible binary to be present
in PATH
*/
func AnsiblePlaybook(playbook string, vars string, user string) {
	cmd0 := "ansible-playbook"
	cmd1 := strings.Trim(playbook, "\n")
	cmd2 := "-i"
	cmd3 := "/root/hosts"
	cmd4 := "--user"
	cmd5 := strings.Trim(user, "\n")
	cmd8 := "--private-key"
	cmd9 := SSH_KEY_PATH
	cmd10 := "--ssh-extra-args"
	cmd11 := "-o StrictHostKeyChecking=no"

	if len(vars) > 0 {
		cmd6 := "-e"
		cmd7 := "@" + REPO_LOCATION + strings.Trim(vars, "\n")
		command := []string{cmd0, cmd1, cmd2, cmd3, cmd4, cmd5, cmd6, cmd7, cmd8, cmd9, cmd10, cmd11}
		ColorPrint(INFO, "Executing: "+strings.Join(command, " "))
		cmd := exec.Command(cmd0, cmd1, cmd2, cmd3, cmd4, cmd5, cmd6, cmd7, cmd8, cmd9, cmd10, cmd11)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		FailError(err)

	} else {
		command := []string{cmd0, cmd1, cmd2, cmd3, cmd4, cmd5, cmd8, cmd9, cmd10, cmd11}
		ColorPrint(INFO, "Executing: "+strings.Join(command, " "))
		cmd := exec.Command(cmd0, cmd1, cmd2, cmd3, cmd4, cmd5, cmd8, cmd9, cmd10, cmd11)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			ColorPrint(ERROR, "%v", err)
		}
	}
}

func generateAnsibleInventory(ipAddr string, ansibleTag string, hostName string, sshUser string) {
	// Assuming SSH port is 22
	d1 := []byte("[" + strings.Trim(ansibleTag, "\n") + "]\n" + hostName + " ansible_host=" + ipAddr + " ansible_port=22 ansible_user=" + sshUser + "\n")
	err := os.WriteFile(INVENTORY_PATH, d1, 0644)
	FailError(err)
}
