package main

import (
	"bytes"
	"log"

	"github.com/Telmate/proxmox-api-go/proxmox"
)

/*
Creates a new clone
of the provided template and
configures it according to cloudInitConfig
*/
func Clone(client *proxmox.Client, template string, cloudInitConfig []byte, node string) *proxmox.VmRef {
	config, err := proxmox.NewConfigQemuFromJson(bytes.NewReader(cloudInitConfig))
	FailError(err)
	log.Println("Looking for template: " + template)
	sourceVmrs, err := client.GetVmRefsByName(template)
	FailError(err)
	if sourceVmrs == nil {
		log.Fatal("Can't find template")
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
	err = proxmox.WaitForShutdown(vmr, client)
	FailError(err)
	log.Println("Completed cloning process")
	vmr = proxmox.NewVmRef(config.VmID)
	return vmr
}

/*
Deletes an existing VM
using its vmid
*/
func Destroy(client *proxmox.Client, vmid int) {
	vmr := proxmox.NewVmRef(vmid)
	jbody, err := client.StopVm(vmr)
	ColorPrint(INFO, jbody)
	FailError(err)
	jbody, err = client.DeleteVm(vmr)
	FailError(err)
	ColorPrint(INFO, jbody)
}

/*
Starts an existing VM
using its vmid
*/
func Start(client *proxmox.Client, vmid int) string {
	vmr := proxmox.NewVmRef(vmid)
	jbody, err := client.StartVm(vmr)
	FailError(err)
	return jbody
}

/*
Stops an existing VM
using its vmid
*/
func Stop(client *proxmox.Client, vmid int) string {
	vmr := proxmox.NewVmRef(vmid)
	jbody, err := client.StopVm(vmr)
	FailError(err)
	return jbody
}
