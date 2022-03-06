package main

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"strconv"
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
func main() {

	// Validate the proxmox setup
	timeout, tlsConf, template, node, cpuLimit, memLimit := validateInputs()
	cloudInitConfig, err := os.ReadFile("/etc/cloud/cloud-init")
	if err != nil {
		log.Fatalf("Cloud-Init config not found")
	}
	client := CreateClient(tlsConf, timeout)

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
		overall_cpu_percentage := 0
		overall_mem_percentage := 0
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

			overall_cpu_percentage += int(used_cpu / total_cpu)
			overall_mem_percentage += int(used_mem.MilliValue() / total_mem.MilliValue())

		}

		overall_cpu_percentage /= nodes.Size()
		overall_mem_percentage /= nodes.Size()
		ColorPrint(INFO, "Overall cpu usage: %d and overall mem usage: %d\n", overall_cpu_percentage, overall_mem_percentage)

		if overall_cpu_percentage > cpuLimit || overall_mem_percentage > memLimit {
			ColorPrint(INFO, "Creating new VM")
			ColorPrint(INFO, "Using the following params: %s , %s , %s, %s", client.ApiUrl, template, cloudInitConfig, node)
			Clone(client, template, cloudInitConfig, node)
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
