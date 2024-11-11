package main

import (
	"bytes"
	"errors"
	"log"
	"time"

	"github.com/Telmate/proxmox-api-go/proxmox"
)

/*

Creates a new clone of the provided template and
configures it according to cloudInitConfig
*/
func CloneVM(client *proxmox.Client, template string, cloudInitConfig []byte, node string, runAnsiblePlaybook bool) (*proxmox.ConfigQemu, *proxmox.VmRef) {
	config, err := proxmox.NewConfigQemuFromJson(bytes.NewReader(cloudInitConfig))
	if runAnsiblePlaybook {
		// Enable qemu agent - needed for ansible
		config.Agent = 1
	}

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
	return config, vmr
}

/*
Destroy function stops and
deletes am existing VM
using its vmid
*/
func DestroyVM(client *proxmox.Client, vmid int) (string, error) {
	vmr := proxmox.NewVmRef(vmid)
	jbody, err := client.StopVm(vmr)
	if err != nil {
		return jbody, err
	}
	ColorPrint(INFO, jbody)
	jbody, err = client.DeleteVm(vmr)
	ColorPrint(INFO, jbody)
	return jbody, err
}

//Starts an existing VM using its vmid
func StartVM(client *proxmox.Client, vmid int) (string, error) {
	vmr := proxmox.NewVmRef(vmid)
	jbody, err := client.StartVm(vmr)
	if err != nil {
		return "", err
	}
	return jbody, err
}

//Stops an existing VM using its vmid
func StopVM(client *proxmox.Client, vmid int) string {
	vmr := proxmox.NewVmRef(vmid)
	jbody, err := client.StopVm(vmr)
	FailError(err)
	return jbody
}

// Wait for VM to power on
func WaitForPowerOn(vmr *proxmox.VmRef, client *proxmox.Client) (err error) {
	for ii := 0; ii < 100; ii++ {
		vmState, err := client.GetVmState(vmr)
		if err != nil {
			log.Print("Wait error:")
			log.Println(err)
		} else if vmState["status"] == "running" {
			return nil
		}
		ColorPrint(INFO, "Waiting for VM to power on.")
		time.Sleep(5 * time.Second)
	}
	return errors.New("VM did not start within wait time")
}

// Wait for ping success from Qemu Agent
func WaitForQemuAgent(vmr *proxmox.VmRef, client *proxmox.Client) (err error) {
	for ii := 0; ii < 100; ii++ {
		_, err := client.QemuAgentPing(vmr)
		if err != nil {
			log.Print("Wait error:")
			log.Println(err)
		} else {
			return nil
		}
		ColorPrint(INFO, "Waiting for qemu agent to start for VM %d.", vmr.VmId())
		time.Sleep(5 * time.Second)
	}
	return errors.New("qemu agent did not start within wait time")
}
