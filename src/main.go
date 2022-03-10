package main

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Telmate/proxmox-api-go/proxmox"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

/*
Uses all other functions defined
in this module to calculate overall
mem and cpu usage. Creates a new VM
if any one of the overall thresholds
are exceeded.
*/

const (
	REPO_LOCATION = "/root/ansible/"
)

func main() {

	fRun := true

	// Validate the proxmox setup
	timeout, tlsConf, template, node, cpuLimit, memLimit := validateInputs()
	cloudInitConfig, err := os.ReadFile("/etc/cloud/cloud-init")
	if err != nil {
		log.Fatalf("Cloud-Init config not found")
	}
	client := CreateClient(tlsConf, timeout)

	// Validate postgres setup
	connStr := validatePostgresConfig()

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	mc, err := metrics.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	for {

		//Get all nodes
		nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		// Loop through all nodes and find allocatable cpu & mem along with usage
		var overall_cpu_percentage, overall_mem_percentage float32 = 0.00, 0.00
		for node_index := range nodes.Items {
			name := nodes.Items[node_index].Name
			total_cpu := nodes.Items[node_index].Status.Allocatable.Cpu().MilliValue()
			total_mem := nodes.Items[node_index].Status.Allocatable.Memory()
			metric_values, err := mc.MetricsV1beta1().NodeMetricses().Get(context.TODO(), name, metav1.GetOptions{})
			if err != nil {
				panic(err.Error())
			}
			used_mem := metric_values.Usage.Memory()
			used_cpu := metric_values.Usage.Cpu().MilliValue()
			ColorPrint(INFO, "Node %s is using %s/%s mem and %d/%d cpu\n", name, used_mem, total_mem, used_cpu, total_cpu)

			overall_cpu_percentage += float32(used_cpu / total_cpu)
			overall_mem_percentage += float32(used_mem.MilliValue() / total_mem.MilliValue())

		}

		overall_cpu_percentage = overall_cpu_percentage / float32(nodes.Size())
		overall_mem_percentage = overall_mem_percentage / float32(nodes.Size())
		ColorPrint(INFO, "Overall cpu usage: %f and overall mem usage: %f\n", overall_cpu_percentage, overall_mem_percentage)

		if overall_cpu_percentage > float32(cpuLimit) || overall_mem_percentage > float32(memLimit) || fRun {
			fRun = !fRun

			// Clone repo for ansible if config is provided
			ansibleTag := getValueOf("ansibleTag", "")
			ansibleRepo := getValueOf("ansibleRepo", "")
			var playbookLocation string
			runAnsiblePlaybook := false
			if len(ansibleTag) != 0 && len(ansibleRepo) != 0 {
				runAnsiblePlaybook = true
				ColorPrint(INFO, "Ansible Tag and Repo were provided in the configuration: %s", ansibleTag)
				ColorPrint(INFO, "Attempting to configure this new VM with the ansible config provided.")
				CloneRepo(ansibleRepo)
				repoSubFolder := getValueOf("repoSubFolder", "")
				if len(repoSubFolder) != 0 {
					playbookLocation = REPO_LOCATION + repoSubFolder
					ColorPrint(INFO, "Using path: '%s' for running ansible-playbook", playbookLocation)
				}
			}

			ColorPrint(INFO, "Creating new VM")
			ColorPrint(INFO, "Using the following params: %s , %s , %s, %s", client.ApiUrl, template, cloudInitConfig, node)
			config, vmr := CloneVM(client, template, cloudInitConfig, node, runAnsiblePlaybook)
			InsertVmInfo(connStr, vmr, config)

			// Start the VM
			ColorPrint(INFO, "Attempting to start the VM")
			res := StartVM(client, vmr.VmId())
			ColorPrint(INFO, res)

			// Wait for qemu agent to come up
			err = WaitForQemuAgent(vmr, client)
			for err != nil {
				ColorPrint(ERROR, "Qemu Agent not running for VM with id: %d on node %s", vmr.VmId(), vmr.Node())
				ColorPrint(WARN, "Attempting to wait for qemu agent.")
				err = WaitForQemuAgent(vmr, client)
			}

			// Wait for VM to attain an IP address
			ColorPrint(WARN, "Waiting for the VM to get an IP Address.")
			time.Sleep(10 * time.Second)

			// Figure out the IP Address assigned to the VM
			var ipAddress string
			interfaces, err := client.GetVmAgentNetworkInterfaces(vmr)
			FailError(err)
			for _, interfaceData := range interfaces {
				for index, ipArrd := range interfaceData.IPAddresses {
					if strings.Contains(interfaceData.Name, "eth") && len(ipAddress) == 0 {
						ipAddress = ipArrd.String()
					}
					ColorPrint(INFO, "FOUND IP ADDRESS: %s for INTERFACE: %s on index %d", ipArrd.String(), interfaceData.Name, index)
				}
			}
			ColorPrint(INFO, "Using %s as the IP Address of the created VM", ipAddress)

			// Run ansible playbook(s)
			if runAnsiblePlaybook {
				// TODO: Do we need any other interface/ipaddress here?
				ExecutePlaybook(playbookLocation, ipAddress, ansibleTag)
			}
		}
		time.Sleep(10 * time.Second)
	}
}

/*
validateInputs validates that all
required inputs are in place and
 are using the correct formats.
*/
func validateInputs() (int, *tls.Config, string, string, int, int) {
	insecure, err := strconv.ParseBool(getValueOf("insecure", "false"))
	FailError(err)
	*proxmox.Debug, err = strconv.ParseBool(getValueOf("debug", "false"))
	FailError(err)
	taskTimeout, err := strconv.Atoi(getValueOf("taskTimeout", "300"))
	FailError(err)
	memLimit := getValueOf("memoryLimit", "")
	if len(memLimit) == 0 {
		log.Fatal("memoryLimit not specified in config!")
	}
	memoryLimit, err := strconv.Atoi(memLimit)
	FailError(err)
	cLimit := getValueOf("cpuLimit", "")
	if len(cLimit) == 0 {
		log.Fatal("cpuLimit not specified in config!")
	}
	cpuLimit, err := strconv.Atoi(cLimit)
	FailError(err)
	node := getValueOf("nodeName", "")
	if len(node) == 0 {
		log.Fatal("Node name not specified in config!")
	}
	template := getValueOf("templateName", "")
	if len(template) == 0 {
		log.Fatal("Template name not specified in config!")
	}
	tlsconf := &tls.Config{InsecureSkipVerify: true}
	if !insecure {
		tlsconf = nil
	}
	return taskTimeout, tlsconf, template, node, cpuLimit, memoryLimit
}
