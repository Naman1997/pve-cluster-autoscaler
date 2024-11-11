package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/relex/aini"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
	REPO_LOCATION       = "/root/repo/"
	INVENTORY_PATH      = "/root/hosts"
	SSH_KEY_PATH        = "/etc/ssh/id_rsa"
	CLOUD_INIT_PATH     = "/etc/cloud/cloud-init"
	RETRY_PERIOD        = 10
	INTERFACE_SUBSTRING = "eth"
)

func main() {

	fRun := true

	// Validate the proxmox setup
	timeout, tlsConf, template, node, cpuLimit, memLimit, joinCommand := validateInputs()
	cloudInitConfig, err := os.ReadFile(CLOUD_INIT_PATH)
  
	if err != nil {
		ColorPrint(ERROR, "Cloud-Init config not found. Error: %v", err)
	}
	client := CreateClient(tlsConf, timeout)

	// Validate postgres setup
	connStr := validatePostgresConfig()

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	FailError(err)
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	FailError(err)

	mc, err := metrics.NewForConfig(config)
	FailError(err)

	for {

		//Get all nodes
		nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		FailError(err)

		// Loop through all nodes and find allocatable cpu & mem along with usage
		var overall_cpu_percentage, overall_mem_percentage float32 = 0.00, 0.00
		for node_index := range nodes.Items {
			name := nodes.Items[node_index].Name
			total_cpu := nodes.Items[node_index].Status.Allocatable.Cpu().MilliValue()
			total_mem := nodes.Items[node_index].Status.Allocatable.Memory()
			metric_values, err := mc.MetricsV1beta1().NodeMetricses().Get(context.TODO(), name, metav1.GetOptions{})
			FailError(err)
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
				FailError(CloneRepo(ansibleRepo))
				ansiblePlaybook := getValueOf("ansiblePlaybook", "")
				if len(ansiblePlaybook) != 0 {
					playbookLocation = REPO_LOCATION + ansiblePlaybook
					ColorPrint(INFO, "Using path: '%s' for running ansible-playbook", playbookLocation)
				}
			}

			ColorPrint(INFO, "Creating new VM...")
			ColorPrint(INFO, "Using the following params: %s , %s , %s, %s", client.ApiUrl, template, cloudInitConfig, node)
			config, vmr := CloneVM(client, template, cloudInitConfig, node, runAnsiblePlaybook)
			InsertVmInfo(connStr, vmr, config)

			// Start the VM
			ColorPrint(INFO, "Attempting to start the VM...")
			res, err := StartVM(client, vmr.VmId())
			ColorPrint(INFO, res)
			for err != nil {
				ColorPrint(WARN, "Encountered an error while trying to start the VM: %v", err)
				ColorPrint(INFO, "Attempting to start the VM again...")
				time.Sleep(RETRY_PERIOD * time.Second)
				res, err = StartVM(client, vmr.VmId())
				ColorPrint(INFO, res)
			}
			err = WaitForPowerOn(vmr, client)
			for err != nil {
				ColorPrint(WARN, "VM did not start up in the expected time period: %v", err)
				ColorPrint(INFO, "Attempting to start the VM again...")
				time.Sleep(RETRY_PERIOD * time.Second)
				StartVM(client, vmr.VmId())
				err = WaitForPowerOn(vmr, client)
			}

			// Wait for qemu agent to come up
			err = WaitForQemuAgent(vmr, client)
			for err != nil {
				ColorPrint(WARN, "Qemu Agent was not running for VM with id: %d on node %s", vmr.VmId(), vmr.Node())
				ColorPrint(WARN, "Attempting to wait for qemu agent...")
				err = WaitForQemuAgent(vmr, client)
			}

			// Wait for VM to attain an IP address
			var ipAddress string
			for len(ipAddress) == 0 {
				ColorPrint(WARN, "Waiting for the VM to get an IP Address...")
				time.Sleep(RETRY_PERIOD * time.Second)

				// Figure out the IP Address assigned to the VM
				interfaces, err := client.GetVmAgentNetworkInterfaces(vmr)
				for err != nil {
					ColorPrint(WARN, "Encountered an error while getting the network interfaces: %v", err)
					ColorPrint(WARN, "Attempting to get the interfaces again...")
					time.Sleep(RETRY_PERIOD * time.Second)
					interfaces, err = client.GetVmAgentNetworkInterfaces(vmr)
				}
				for _, interfaceData := range interfaces {
					for index, ipArrd := range interfaceData.IPAddresses {
						// Assuming interface name containse eth
						if strings.Contains(interfaceData.Name, INTERFACE_SUBSTRING) && len(ipAddress) == 0 {
							ipAddress = ipArrd.String()
						}
						ColorPrint(INFO, "FOUND IP ADDRESS: %s for INTERFACE: %s on index %d", ipArrd.String(), interfaceData.Name, index)
					}
				}
			}
			ColorPrint(INFO, "Using %s as the IP Address of the created VM", ipAddress)

			// Run ansible playbook(s)
			sshUser := getValueOf("sshUser", "admin")
			if runAnsiblePlaybook {
				generateAnsibleInventory(ipAddress, ansibleTag, config.Name, sshUser)
				AnsibleGalaxy(REPO_LOCATION + getValueOf("ansibleRequirements", ""))
				ColorPrint(INFO, "Generating ansible inventory...")
				time.Sleep(2 * time.Second)

				// Parse the inventory
				file, err := os.Open(INVENTORY_PATH)
				FailError(err)
				inventoryReader := bufio.NewReader(file)
				_, err = aini.Parse(inventoryReader)
				for err != nil {
					ColorPrint(WARN, "There might be an issue with the provided params")
					ColorPrint(INFO, "Re-generating ansible inventory...")
					ColorPrint(INFO, "Params provided: [IP ADDRESS: %s] [ANSIBLE TAG: %s] [HOSTNAME: %s] [SSH USER: %s]", ipAddress, ansibleTag, config.Name, sshUser)
					generateAnsibleInventory(ipAddress, ansibleTag, config.Name, sshUser)
					time.Sleep(2 * time.Second)
					inventoryReader = bufio.NewReader(file)
					_, err = aini.Parse(inventoryReader)
				}

				// Run the playbook provided
				err = AnsiblePlaybook(playbookLocation, getValueOf("ansibleExtraVarsFile", ""), sshUser, joinCommand)
				// Retry once on failure
				if err != nil {
					ColorPrint(WARN, "An error occoured while running the ansible playbook: %v", err)
					ColorPrint(WARN, "Retrying once more to run the ansible playbook....")
					AnsiblePlaybook(playbookLocation, getValueOf("ansibleExtraVarsFile", ""), sshUser, joinCommand)
				}
				if err != nil {
					ColorPrint(WARN, "Errors encountered while running the playbook: %v", err)
					ColorPrint(WARN, "Node with IP: '%s' and ID: '%d' was unable to join the cluster!", ipAddress, vmr.VmId())
					ColorPrint(WARN, "Deleting the VM with IP: '%s' and ID: '%d'", ipAddress, vmr.VmId())
					res, err = DestroyVM(client, vmr.VmId())
					for err != nil {
						ColorPrint(WARN, "%v", err)
						res, err = DestroyVM(client, vmr.VmId())
					}
					ColorPrint(INFO, res)
				}
			} else {
				err = sendCommands(sshUser, ipAddress, joinCommand)
				if err != nil {
					ColorPrint(WARN, "Invalid Join Command: Expired token?")
					ColorPrint(WARN, "Node with IP: '%s' and ID: '%d' was unable to join the cluster!", ipAddress, vmr.VmId())
					ColorPrint(WARN, "Deleting the VM with IP: '%s' and ID: '%d'", ipAddress, vmr.VmId())
					res, err = DestroyVM(client, vmr.VmId())
					for err != nil {
						ColorPrint(WARN, "%v", err)
						res, err = DestroyVM(client, vmr.VmId())
					}
					ColorPrint(INFO, res)
				}
			}

			// Attempt to add worker role the newly created node
			payload := `{"metadata":{"labels":{"kubernetes.io/role":"worker"}}}`
			response, err := clientset.
				CoreV1().
				Nodes().
				Patch(context.TODO(),
					config.Name,
					types.MergePatchType,
					[]byte(payload),
					metav1.PatchOptions{FieldManager: "kubectl-label"})

			ColorPrint(INFO, response.String())
			ColorPrint(WARN, "Errors for labeling new node: %v", err)

			// TODOs : https://github.com/users/Naman1997/projects/6
			for {
				time.Sleep(RETRY_PERIOD * time.Second)
				ColorPrint(INFO, "Reached TODO point")
			}

		}
		time.Sleep(RETRY_PERIOD * time.Second)
	}
}

/*
validateInputs validates that all
required inputs are in place and
 are using the correct formats.
*/
func validateInputs() (int, *tls.Config, string, string, int, int, string) {
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
	joinCommand := getValueOf("joinCommand", "")
	if len(joinCommand) == 0 {
		log.Fatal("joinCommand not specified in config!")
	}
	return taskTimeout, tlsconf, template, node, cpuLimit, memoryLimit, joinCommand
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
	memoryLimit, err := strconv.Atoi(getValueOf("memoryLimit", ""))
	FailError(err)
	cpuLimit, err := strconv.Atoi(getValueOf("cpuLimit", ""))
	FailError(err)
	node := getValueOf("nodeName", "")
	if node == "" {
		log.Fatal("Node not specified!")
	}
	template := getValueOf("templateName", "")
	if template == "" {
		log.Fatal("Template not specified!")
	}
	tlsconf := &tls.Config{InsecureSkipVerify: true}
	if !insecure {
		tlsconf = nil
	}
	return taskTimeout, tlsconf, template, node, cpuLimit, memoryLimit
}
