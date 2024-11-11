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
		FailError(err)

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
func AnsiblePlaybook(playbook string, vars string, user string, joinCommand string) error {
	cmd0 := "ansible-playbook"
	cmd1 := strings.Trim(playbook, "\n")
	cmd2 := "-i"
	cmd3 := "/root/hosts"
	cmd4 := "--user"
	cmd5 := strings.Trim(user, "\n")
	cmd6 := "--private-key"
	cmd7 := SSH_KEY_PATH
	cmd8 := "--ssh-extra-args"
	cmd9 := "-o StrictHostKeyChecking=no"
	cmd10 := "-e"

	var cmd *exec.Cmd
	command := []string{cmd0, cmd1, cmd2, cmd3, cmd4, cmd5, cmd8, cmd9, cmd6, cmd7}
	if len(vars) > 0 {
		cmd11 := "@" + REPO_LOCATION + strings.Trim(vars, "\n")
		command = append(command, cmd10, cmd11)
	}

	/*
		Attempt to both inject join-command as an env var and create it as a
		file in case user is using something like:
		```
		- name: Copy the join command to server location
			copy: src=join-command dest=/tmp/join-command.sh mode=0777
		- name: Join the node to cluster
			command: sh /tmp/join-command.sh
		```
	*/
	if len(joinCommand) != 0 {
		cmd12 := "'join-command=" + strings.Trim(joinCommand, "\n") + "'"
		command = append(command, cmd10, cmd12)
		generateJoinFile(joinCommand, playbook[:+strings.LastIndex(playbook, "/")+1])
	}

	ColorPrint(INFO, "Executing: "+strings.Join(command, " "))
	cmd = exec.Command(command[0], command[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func generateAnsibleInventory(ipAddr string, ansibleTag string, hostName string, sshUser string) {
	// Assuming SSH port is 22
	d1 := []byte("[" + strings.Trim(ansibleTag, "\n") + "]\n" + hostName + " ansible_host=" + ipAddr + " ansible_port=22 ansible_user=" + sshUser + "\n")
	err := os.WriteFile(INVENTORY_PATH, d1, 0644)
	FailError(err)
}

func generateJoinFile(joinCommand string, folderPath string) {
	d1 := []byte(strings.Trim(joinCommand, "\n"))
	err := os.WriteFile(folderPath+"join-command", d1, 0644)
	FailError(err)
}
