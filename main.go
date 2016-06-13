package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bonclay7/ecs-elbless/elbless"
)

func init() {
	// Change the device for logging to stdout.
	log.SetOutput(os.Stdout)
}

func main() {

	cluster := flag.String("cluster", "", "ECS Cluster ID [mandatory]")
	region := flag.String("region", "", "AWS_DEFAULT_REGION [optional]")
	serviceFilter := flag.String("service-filter", "*", "ECS service id [optional]")

	flag.Usage = func() {
		fmt.Printf("usage: ecs-elbless -cluster <cluster-id> [-service-filter 'filter-exp'] [-region us-west-1]\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *cluster == "" {
		log.Fatal("Error: missing cluster id. To get help, use [--help | -h] option.\n\n")
	}

	// get region from AWS_DEFAULT_REGION
	if *region == "" {
		*region = os.Getenv("AWS_DEFAULT_REGION")
		if *region == "" {
			log.Fatal("Error: missing -region option and AWS_DEFAULT_REGION \n\n")
		}
	}

	printEcsServices(*cluster, *region, *serviceFilter)
}

func printEcsServices(cluster string, region string, filter string) {
	services := elbless.GetServicesEndpoints(cluster, region, filter)

	for service, metadata := range services {
		fmt.Println("Service : ", service)

		for _, m := range metadata {

			fmt.Println("  Task ID: ", m.Task.Task)
			fmt.Println("  Container ID: ", m.Task.Container)
			fmt.Printf("  Public endpoint: %s:%d\n", m.Ec2Infos.PublicIP, m.Task.HostPort)
			fmt.Printf("  Public DNS endpoint: %s:%d\n", m.Ec2Infos.PublicDNSName, m.Task.HostPort)
			fmt.Printf("  Private endpoint: %s:%d\n", m.Ec2Infos.PrivateIP, m.Task.HostPort)
			fmt.Printf("  Private DNS endpoint: %s:%d\n", m.Ec2Infos.PrivateDNSName, m.Task.HostPort)
			fmt.Println("")
		}
	}

}
