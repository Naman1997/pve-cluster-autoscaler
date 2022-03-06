package main

import (
	"bytes"
	"log"

	"github.com/Telmate/proxmox-api-go/proxmox"
)

func Clone(client *proxmox.Client, template string, cloudInitConfig []byte, node string) {
	config, err := proxmox.NewConfigQemuFromJson(bytes.NewReader(cloudInitConfig))
	FailError(err)
	log.Println("Looking for template: " + template)
	sourceVmrs, err := client.GetVmRefsByName(template)
	FailError(err)
	if sourceVmrs == nil {
		log.Fatal("Can't find template")
		return
	}
	vmid, err := client.GetNextID(0)
	FailError(err)
	vmr := proxmox.NewVmRef(vmid)
	vmr.SetNode(node)
	log.Print("Creating node: ")
	log.Println(vmr)
	// prefer source Vm located on same node
	sourceVmr := sourceVmrs[0]
	for _, candVmr := range sourceVmrs {
		if candVmr.Node() == vmr.Node() {
			sourceVmr = candVmr
		}
	}

	FailError(config.CloneVm(sourceVmr, vmr, client))
	FailError(config.UpdateConfig(vmr, client))
	log.Println("Complete")
}
