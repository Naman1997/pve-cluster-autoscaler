package main

import (
	"crypto/tls"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/Telmate/proxmox-api-go/proxmox"
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

/*
CreateClient is used to create
a new client using the provided
tls config and timeout params and
the credentials provided to the pod
via a secret as env vars.
Proxy is currently not being supported.
*/
func CreateClient(tlsconf *tls.Config, taskTimeout int) (client *proxmox.Client) {
	c, err := proxmox.NewClient(getValueOf("PM_API_URL", ""), nil, tlsconf, "", taskTimeout)
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
