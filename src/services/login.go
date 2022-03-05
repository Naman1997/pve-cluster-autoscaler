package services

import (
	"crypto/tls"
	"log"

	"github.com/Telmate/proxmox-api-go/proxmox"
)

func CreateClient(tlsconf *tls.Config, proxyUrl string, taskTimeout int) (client *proxmox.Client) {
	c, err := proxmox.NewClient(getValueOf("PM_API_URL", ""), nil, tlsconf, proxyUrl, taskTimeout)
	FailError(err)
	if userRequiresAPIToken(getValueOf("PM_USER", "")) {
		c.SetAPIToken(getValueOf("PM_USER", ""), getValueOf("PM_PASS", ""))
		// As test, get the version of the server
		_, err := c.GetVersion()
		if err != nil {
			log.Fatalf("login error: %s", err)
		}
	} else {
		err = c.Login(getValueOf("PM_USER", ""), getValueOf("PM_PASS", ""), getValueOf("PM_OTP", ""))
		FailError(err)
	}
	return c
}
