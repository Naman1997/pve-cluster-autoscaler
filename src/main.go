package main

// "context"
// "fmt"
// "time"

// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// "k8s.io/client-go/kubernetes"
// "k8s.io/client-go/rest"
// metrics "k8s.io/metrics/pkg/client/clientset/versioned"

import (
	"log"
	"time"

	"github.com/Naman1997/pve-cluster-autoscaler/services"
)

func main() {
	config := services.CreateVMClone()
	for {

		log.Println("CPU: ", config.QemuCpu)
		log.Println("Vcpus: ", config.QemuVcpus)
		log.Println("Memory: ", config.Memory)
		log.Println("Storage", config.Storage)
		log.Println("Sockets", config.QemuSockets)
		log.Println("FullClone", config.FullClone)
		log.Println("Nameserver", config.Nameserver)
		log.Println("QemuKVM", config.QemuKVM)
		time.Sleep(10 * time.Second)
	}
	// // creates the in-cluster config
	// config, err := rest.InClusterConfig()
	// if err != nil {
	// 	panic(err.Error())
	// }
	// // creates the clientset
	// clientset, err := kubernetes.NewForConfig(config)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// mc, err := metrics.NewForConfig(config)
	// if err != nil {
	// 	panic(err)
	// }

	// for {

	// 	//Get all nodes
	// 	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}

	// 	// Loop through all nodes and find allocatable cpu & mem along with usage
	// 	for node_index := range nodes.Items {
	// 		name := nodes.Items[node_index].Name
	// 		total_cpu := nodes.Items[node_index].Status.Allocatable.Cpu().MilliValue()
	// 		total_mem := nodes.Items[node_index].Status.Allocatable.Memory()
	// 		metric_values, err := mc.MetricsV1beta1().NodeMetricses().Get(context.TODO(), name, metav1.GetOptions{})
	// 		if err != nil {
	// 			panic(err.Error())
	// 		}
	// 		used_mem := metric_values.Usage.Memory()
	// 		used_cpu := metric_values.Usage.Cpu().MilliValue()
	// 		fmt.Printf("Node %s is using %s/%s mem and %d/%d cpu\n", name, used_mem, total_mem, used_cpu, total_cpu)
	// 	}

	// 	time.Sleep(10 * time.Second)
	// }
}
