package services

import (
	"bytes"
	"crypto/tls"
	"log"
	"os"
	"strconv"

	"github.com/Telmate/proxmox-api-go/proxmox"
)

func CreateVMClone() *proxmox.ConfigQemu {
	insecure, err := strconv.ParseBool(getValueOf("insecure", "false"))
	FailError(err)
	*proxmox.Debug, err = strconv.ParseBool(getValueOf("debug", "false"))
	FailError(err)
	taskTimeout, err := strconv.Atoi(getValueOf("taskTimeout", "300"))
	FailError(err)
	proxyUrl := getValueOf("proxyUrl", "")
	nodeName := getValueOf("nodeName", "")
	if nodeName == "" {
		log.Fatal("Node not specified!")
	}
	templateName := getValueOf("templateName", "")
	if templateName == "" {
		log.Fatal("Template not specified!")
	}
	tlsconf := &tls.Config{InsecureSkipVerify: true}
	if !insecure {
		tlsconf = nil
	}

	c := CreateClient(tlsconf, proxyUrl, taskTimeout)
	value, err := os.ReadFile("/etc/secrets/cloud-init")
	if err != nil {
		log.Fatalf("Cloud-Init config not found")
	}
	config, err := proxmox.NewConfigQemuFromJson(bytes.NewReader(value))
	FailError(err)
	log.Println("Looking for template: " + templateName)
	sourceVmrs, err := c.GetVmRefsByName(templateName)
	FailError(err)
	if sourceVmrs == nil {
		log.Fatal("Can't find template")
	}

	vmid, err := c.GetNextID(0)
	FailError(err)
	vmr := proxmox.NewVmRef(vmid)
	vmr.SetNode(nodeName)
	log.Print("Creating node: ")
	log.Println(vmr)
	// prefer source Vm located on same node
	sourceVmr := sourceVmrs[0]
	for _, candVmr := range sourceVmrs {
		if candVmr.Node() == vmr.Node() {
			sourceVmr = candVmr
		}
	}

	FailError(config.CloneVm(sourceVmr, vmr, c))
	FailError(config.UpdateConfig(vmr, c))
	log.Println("Complete")

	return config
}

// func validateInputs() (string, string, string, int, int, int, string, string, string, int) {
// 	name := getValueOf("name", "AutoScaled kworker")
// 	desc := getValueOf("desc", "")
// 	storage := getValueOf("storage", "")
// 	mem, err := strconv.Atoi(getValueOf("memory", ""))
// 	FailError(err)
// 	cpu, err := strconv.Atoi(getValueOf("cores", ""))
// 	FailError(err)
// 	sockets, err := strconv.Atoi(getValueOf("sockets", ""))
// 	FailError(err)
// 	bridge := getValueOf("bridge", "")
// 	sshKey := getValueOf("sshKey", "")
// 	nameserver := getValueOf("nameserver", "")
// 	balloon, err := strconv.Atoi(getValueOf("balloon", "0"))
// 	FailError(err)

// 	if len(storage) == 0 {
// 		log.Fatal("'storage' not provided for VM in config. Example value: ''")
// 	}
// 	if mem == 0 {
// 		log.Fatal("'memory' not provided for VM in config")
// 	}
// 	if cpu == 0 {
// 		log.Fatal("'cores' not provided for VM in config")
// 	}
// 	if sockets == 0 {
// 		log.Fatal("'sockets' not provided for VM in config")
// 	}

// 	return name, desc, storage, mem, cpu, sockets, bridge, sshKey, nameserver, balloon
// }

// func createNewConfig() *proxmox.ConfigQemu {
// 	name, desc, storage, mem, cpu, sockets, bridge, sshKey, nameserver, balloon := validateInputs()
// 	config := &proxmox.ConfigQemu{QemuVlanTag: -1, QemuKVM: true}
// 	config.Name = name
// 	config.Description = desc
// 	config.Storage = storage
// 	config.Memory = mem
// 	config.QemuCores = cpu
// 	config.QemuVcpus = cpu * sockets
// 	config.QemuSockets = sockets
// 	config.QemuBrige = bridge
// 	config.Sshkeys = sshKey
// 	config.Nameserver = nameserver
// 	config.Balloon = balloon
// 	config.Agent = 1
// 	config.FullClone = createPtr(1)
// 	return config
// }

// func createPtr(x int) *int {
// 	return &x
// }

// name=kworker
// desc=Worker Node Generated By AutoScaler
// storage=local-lvm
// memory=4096
// cores=4
// sockets=1
// bridge=vmbr0
// sshkeys=ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDCNwzceEh+KtHHxcKKRcLTGzl5VoClnHNJ0BKk5SyLcuKZilyqLKEYcEDS+BbeMFrFaQvuF7bkr/bS8ruhIE/0Jl789QuWrACz7QS78/43/rondyjDB9P6n6dXSBvN3T/Pa62XeMj6XPOZ2SfYya1ZMjIgYT+6LS4tpxnkPAdcCKaIGdI+LjrPSNFgNZ5CmgcreulaWkeuN2kGg3f0QmaBBZAZF3xarwtA7pU7mNS/YIOGDFOgEGweu7upmq7Ps9f+nWcxW3CjHM0nGldeX3MgLq+AfDqjgi1xk64AHnPY2iKpfbMdpqJKjRLt+DBlau6D8xcg8itoAAN2HeJXkzQKqGnAOQ9RFZq38yopMtrj6YkiCHzPoLK1NGHIxgWAH+3MdLdfAHYzGAJwIqL7axfsfQcU7hs/55r1aQ9lPtI2e7fm1Y34yXo8Jp24D/q9oRQM6zWj22hwzNwvHq/QZOgSY3CPvxfbaWt/SEbId2oMhzrrjknmd+/Kha9jTyiZWv+4Xm1qwag+PEKtrTBwlMRxwiytrLuYVQC4dFXMLJEvY22k8p/aukOMdLe2ZgQOXj4kYmDdAlYdbylx+5fUHGAbD4ZkQAdHkNXWgG7wXqxdEcPbZliYiWnPUUhp14l/4XuRFGjz3L7IxCy1Rf40u4NPaVm+wIWKjiwGNG40XZN7vQ== naman@naman.dev.com",
// nameserver=8.8.8.8
// PM_API_URL=https://192.168.0.106:8006/api2/json
// PM_PASS=united@123
// PM_USER=root@pam
// PM_OTP=
// insecure=true
// debug=false
// nodeName=loki
// templateName=arch-template
// taskTimeout=300
// balloon=0
// cloud-init=

// resource "proxmox_vm_qemu" "kworker" {
// 	count                     = var.kworker_config.count
// 	name                      = format("kworker%s", count.index)
// 	desc                      = "Worker node in k8s cluster."
// 	os_type                   = "cloud-init"
// 	clone                     = var.CLONE_TEMPLATE
// 	fullclone                = true
// 	agent                     = var.common_configs.agent
// 	target_node               = var.PROXMOX_NODE
// 	onboot                    = var.kworker_config.onboot
// 	memory                    = var.kworker_config.memory
// 	sockets                   = var.kworker_config.sockets
// 	cores                     = var.kworker_config.cores
// 	nameserver                = var.NAMESERVER
// 	boot                      = var.BOOT_ORDER

// 	network {
// 	  model  = var.common_configs.network_model
// 	  bridge = var.DEFAULT_BRIDGE
// 	}

//   }
