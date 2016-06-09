package elbless

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
)

// TaskWrapper contains useful values from ecs describe-tasks call
type TaskWrapper struct {
	serviceName       string
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

var sess = session.New()

// Fetch from AWS tasks created for an ECS cluster
func fetchTasksIDs(clusterID string, region string) []string {
	svc := ecs.New(sess, &aws.Config{Region: aws.String(region)})

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

func fetchTaskDescription(clusterID string, taskID string, region string) *ecs.Task {
	svc := ecs.New(sess, &aws.Config{Region: aws.String(region)})

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

func filterTasks(clusterID string, tasks []string, region string) []TaskWrapper {

	slice := make([]TaskWrapper, 0, len(tasks))

	for _, task := range tasks {
		taskDescription := fetchTaskDescription(clusterID, task, region)

		newTaskWrapper := new(TaskWrapper)
		newTaskWrapper.serviceName = strings.ToLower(*taskDescription.Containers[0].Name)
		newTaskWrapper.container = strings.Split(*taskDescription.Containers[0].ContainerArn, "/")[1]
		newTaskWrapper.containerInstance = strings.Split(*taskDescription.ContainerInstanceArn, "/")[1]
		newTaskWrapper.task = task
		newTaskWrapper.hostPort = *taskDescription.Containers[0].NetworkBindings[0].HostPort

		slice = append(slice, *newTaskWrapper)
	}

	return slice
}

func fetchContainerInstance(clusterID string, task TaskWrapper, region string) string {
	svc := ecs.New(sess, &aws.Config{Region: aws.String(region)})

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

func fetchEC2Instance(instanceID string, region string) EC2Wrapper {
	svc := ec2.New(sess, &aws.Config{Region: aws.String(region)})

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

func getMicroservices(clusterID string, tasks []TaskWrapper, region string) (MicroservicesMap map[string][]Microservice) {

	MicroservicesMap = make(map[string][]Microservice)

	for _, task := range tasks {
		containerEC2InstanceID := fetchContainerInstance(clusterID, task, region)
		ec2Instance := fetchEC2Instance(containerEC2InstanceID, region)

		newMicroservice := new(Microservice)
		newMicroservice.ec2Infos = ec2Instance
		newMicroservice.task = task

		MicroservicesMap[task.serviceName] = append(MicroservicesMap[task.serviceName], *newMicroservice)
	}

	return MicroservicesMap

}

// GetServicesEndpoints returns ecs containers endpoints
func GetServicesEndpoints(clusterID string, region string) (MicroservicesMap map[string][]Microservice) {

	// Retrive all the tasks
	tasksIDs := fetchTasksIDs(clusterID, region)

	//Filter for tasks matching our serviceID
	tasks := filterTasks(clusterID, tasksIDs, region)

	microservicesMap := getMicroservices(clusterID, tasks, region)

	return microservicesMap
}
