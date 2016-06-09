package main

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
)

// TaskWrapper contains useful values from ecs describe-tasks call
type TaskWrapper struct {
	container         string
	containerInstance string
	task              string
	hostPort          int64
}

// EC2Wrapper contains useful values from ec2 describe-instances call
type EC2Wrapper struct {
	privateIP      string
	publicIP       string
	privateDNSName string
	publicDNSName  string
}

// Microservice is a full definition of an ecs container
type Microservice struct {
	task     TaskWrapper
	ec2Infos EC2Wrapper
}

const clusterID = "ecs-discovery-ECSCluster-18DVGRRIKGISF"
const serviceID = "GoodReadsApp"
const defaultRegion = "eu-west-1"

var sess = session.New()

// Fetch from AWS tasks created for an ECS cluster
func fetchTasksIDs(clusterID string) []string {
	svc := ecs.New(sess, &aws.Config{Region: aws.String(defaultRegion)})

	params := &ecs.ListTasksInput{
		Cluster: aws.String(clusterID),
	}

	resp, err := svc.ListTasks(params)

	if err != nil {
		panic(err)
	}

	tasksSlice := make([]string, len(resp.TaskArns))

	for idx, task := range resp.TaskArns {
		tasksSlice[idx] = strings.Split(*task, "/")[1]
	}

	return tasksSlice
}

func fetchTaskDescription(clusterID string, taskID string) *ecs.Task {
	svc := ecs.New(sess, &aws.Config{Region: aws.String(defaultRegion)})

	params := &ecs.DescribeTasksInput{
		Tasks: []*string{
			aws.String(taskID),
		},
		Cluster: aws.String(clusterID),
	}

	resp, err := svc.DescribeTasks(params)

	if err != nil {
		panic(err)
	}

	if len(resp.Tasks) == 0 {
		return nil
	}

	if len(resp.Tasks[0].Containers) == 0 {
		return nil
	}

	return resp.Tasks[0]
}

func filterTasks(clusterID string, tasks []string, serviceID string) []TaskWrapper {

	slice := make([]TaskWrapper, 0, len(tasks))

	for _, task := range tasks {
		taskDescription := fetchTaskDescription(clusterID, task)
		// Check for only one container
		if strings.ToLower(*taskDescription.Containers[0].Name) == strings.ToLower(serviceID) {
			newTaskWrapper := new(TaskWrapper)

			newTaskWrapper.container = strings.Split(*taskDescription.Containers[0].ContainerArn, "/")[1]
			newTaskWrapper.containerInstance = strings.Split(*taskDescription.ContainerInstanceArn, "/")[1]
			newTaskWrapper.task = task
			newTaskWrapper.hostPort = *taskDescription.Containers[0].NetworkBindings[0].HostPort

			slice = append(slice, *newTaskWrapper)
		}
	}

	return slice
}

func fetchContainerInstance(clusterID string, task TaskWrapper) string {
	svc := ecs.New(sess, &aws.Config{Region: aws.String(defaultRegion)})

	params := &ecs.DescribeContainerInstancesInput{
		ContainerInstances: []*string{ // Required
			aws.String(task.containerInstance), // Required
			// More values...
		},
		Cluster: aws.String(clusterID),
	}

	resp, err := svc.DescribeContainerInstances(params)

	if err != nil {
		panic(err)
	}

	return *resp.ContainerInstances[0].Ec2InstanceId
}

func fetchEC2Instance(instanceID string) EC2Wrapper {
	svc := ec2.New(sess, &aws.Config{Region: aws.String(defaultRegion)})

	params := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceID), // Required
			// More values...
		},
	}

	resp, err := svc.DescribeInstances(params)

	if err != nil {
		panic(err)
	}

	newEC2Wrapper := new(EC2Wrapper)

	newEC2Wrapper.privateDNSName = *resp.Reservations[0].Instances[0].PrivateDnsName
	newEC2Wrapper.publicDNSName = *resp.Reservations[0].Instances[0].PublicDnsName
	newEC2Wrapper.privateIP = *resp.Reservations[0].Instances[0].PrivateIpAddress
	newEC2Wrapper.publicIP = *resp.Reservations[0].Instances[0].PublicIpAddress

	return *newEC2Wrapper
}

func getMicroservices(clusterID string, tasks []TaskWrapper) []Microservice {

	slice := make([]Microservice, 0, len(tasks))

	for _, task := range tasks {
		containerEC2InstanceID := fetchContainerInstance(clusterID, task)
		ec2Instance := fetchEC2Instance(containerEC2InstanceID)

		newMicroservice := new(Microservice)
		newMicroservice.ec2Infos = ec2Instance
		newMicroservice.task = task

		slice = append(slice, *newMicroservice)
	}

	return slice

}

func main() {
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

}
