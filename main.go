package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bonclay7/ecs-elbless/elbless"
)

func main() {

	cluster := flag.String("cluster", "", "ECS Cluster ID [mandatory]")
	region := flag.String("region", "", "AWS_DEFAULT_REGION [optional]")
	//serviceFilter := flag.String("service-filter", "", "ECS service id [optional]")

	flag.Usage = func() {
		fmt.Printf("usage: ecs-elbless --cluster <cluster-id> [--service-filter filter-exp] [--region us-west-1]\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *cluster == "" {
		fmt.Printf("Error: missing cluster id. To get help, use [--help | -h] option.\n\n")
		os.Exit(1)
	}

	// get region from AWS_DEFAULT_REGION
	if *region == "" {
		*region = os.Getenv("AWS_DEFAULT_REGION")
		if *region == "" {
			fmt.Printf("Error: missing --region option and AWS_DEFAULT_REGION \n\n")
			os.Exit(1)
		}
	}

	fmt.Println(elbless.GetServicesEndpoints(*cluster, *region))
	/*
		// Retrive all the tasks
		tasksIDs := fetchTasksIDs(clusterID)

		//Filter for tasks matching our serviceID
		tasks := filterTasks(clusterID, tasksIDs, serviceID)

		microservices := getMicroservices(clusterID, tasks)

		fmt.Println("clusterID: ", clusterID)

		for _, m := range microservices {
			fmt.Println("  Task ID: ", m.task.task)
			fmt.Println("  Container ID: ", m.task.container)
			fmt.Printf("  Public endpoint: %s:%d\n", m.ec2Infos.publicIP, m.task.hostPort)
			fmt.Printf("  Public DNS endpoint: %s:%d\n", m.ec2Infos.publicDNSName, m.task.hostPort)
			fmt.Printf("  Private endpoint: %s:%d\n", m.ec2Infos.privateIP, m.task.hostPort)
			fmt.Printf("  Private DNS endpoint: %s:%d\n", m.ec2Infos.privateDNSName, m.task.hostPort)
			fmt.Println("")

		}
		//*/
}

func printHelp() {
	fmt.Println("usage: ecs-elbless --cluster <cluster-id> [--service-filter filter-exp] [--region us-west-1]")
	fmt.Println("  --cluster, -c    : ECS Cluster ID")
	fmt.Println("  --service-filter, -s 		: ECS Service ID")
	fmt.Println("  --region, -r     : AWS region : optional")
}
