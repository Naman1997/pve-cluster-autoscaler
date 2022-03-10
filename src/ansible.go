package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/apenella/go-ansible/pkg/options"
	"github.com/apenella/go-ansible/pkg/playbook"
)

func ExecutePlaybook(playbookLocation string, ipAddress string, ansibleTag string) {
	ColorPrint(INFO, "IP ADDRESS: %s", ipAddress)
	ColorPrint(INFO, "PLAYBOOK LOCATION: %s", playbookLocation)
	generateAnsibleInventory(ipAddress, ansibleTag)
	ansiblePlaybookConnectionOptions := &options.AnsibleConnectionOptions{
		User:       getValueOf("sshUser", "admin"),
		PrivateKey: "/etc/ssh/id_rsa",
	}

	ansiblePlaybookOptions := &playbook.AnsiblePlaybookOptions{
		Inventory: "/root/hosts",
		Tags:      "workers",
		Verbose:   true,
	}
	playbook := &playbook.AnsiblePlaybookCmd{
		Playbooks:         []string{playbookLocation},
		ConnectionOptions: ansiblePlaybookConnectionOptions,
		Options:           ansiblePlaybookOptions,
	}

	// TODO : Fix playbook not found error
	// TODO : Figure out how to fix ssh private file permissions inside the container
	playbook.Run(context.TODO())
	// if err != nil {
	// 	panic(err)
	// }
	///usr/local/bin/ansible-playbook --inventory /root/hosts --tags workers -vvvv --private-key /etc/ssh/id_rsa --user naman /root/ansible/playbooks/upgrade.yaml

	for {
		ColorPrint(INFO, "Reached TODO point!")
		time.Sleep(10 * time.Second)
	}
}

func generateAnsibleInventory(ipAddr string, ansibleTag string) {
	d1 := []byte("[" + strings.Trim(ansibleTag, "\n") + "]\n" + ipAddr)
	err := os.WriteFile("/root/hosts", d1, 0644)
	FailError(err)
}
