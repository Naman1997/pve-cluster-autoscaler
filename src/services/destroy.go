package services

import (
	"crypto/tls"
	"log"
	"strconv"

	"github.com/Telmate/proxmox-api-go/proxmox"
)

func DestroyVM(vmid int) {
	var jbody interface{}
	insecure, err := strconv.ParseBool(getValueOf("insecure", "false"))
	FailError(err)
	taskTimeout, err := strconv.Atoi(getValueOf("taskTimeout", "300"))
	FailError(err)
	proxyUrl := getValueOf("proxyUrl", "")
	tlsconf := &tls.Config{InsecureSkipVerify: true}
	if !insecure {
		tlsconf = nil
	}
	c := CreateClient(tlsconf, proxyUrl, taskTimeout)
	vmr := proxmox.NewVmRef(vmid)
	config, err := c.GetVmConfig(vmr)
	FailError(err)
	name := config["name"]
	jbody, err = c.StopVm(vmr)
	FailError(err)
	jbody, err = c.DeleteVm(vmr)
	FailError(err)
	log.Println(jbody)

	log.Printf("Deleted VM with name %s and id %d", name, vmid)
}
