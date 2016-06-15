package elbless

import (
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/gobwas/glob"
)

// TaskWrapper contains useful values from ecs describe-tasks call
type TaskWrapper struct {
	ServiceName       string
	Container         string
	ContainerInstance string
	Task              string
	HostPort          int64
}

// EC2Wrapper contains useful values from ec2 describe-instances call
type EC2Wrapper struct {
	PrivateIP      string
	PublicIP       string
	PrivateDNSName string
	PublicDNSName  string
}

// Microservice is a full definition of an ecs container
type Microservice struct {
	Task     TaskWrapper
	Ec2Infos EC2Wrapper
}

//struct to build the map of microservces after
type mapHelper struct {
	serviceName  string
	microservice Microservice
}

//aws clis
var ecscli = AWSECSClient{nil}
var ec2cli = AWSEC2Client{nil}

// Fetch from AWS tasks created for an ECS cluster
func fetchTasksIDs(clusterID string) ([]string, error) {

	params := &ecs.ListTasksInput{
		Cluster: aws.String(clusterID),
	}

	resp, err := ecscli.ListTasks(params)

	if err != nil {
		return nil, err
	}

	tasksSlice := make([]string, len(resp.TaskArns))

	for idx, task := range resp.TaskArns {
		tasksSlice[idx] = strings.Split(*task, "/")[1]
	}

	return tasksSlice, nil
}

func fetchTaskDescription(clusterID string, taskID string) (*ecs.Task, error) {

	params := &ecs.DescribeTasksInput{
		Tasks: []*string{
			aws.String(taskID),
		},
		Cluster: aws.String(clusterID),
	}

	resp, err := ecscli.DescribeTasks(params)

	if err != nil {
		return nil, err
	}

	if len(resp.Tasks) == 0 {
		err := fmt.Errorf("Task description for %s not found", taskID)
		return nil, err
	}

	if len(resp.Tasks[0].Containers) == 0 {
		err := fmt.Errorf("No containers found in %s", taskID)
		return nil, err
	}

	return resp.Tasks[0], nil
}

func filterTasks(clusterID string, tasks []string, filter string) ([]TaskWrapper, error) {

	var g glob.Glob

	slice := make([]TaskWrapper, 0, len(tasks))
	g = glob.MustCompile(strings.ToLower(filter))

	for _, task := range tasks {
		taskDescription, _ := fetchTaskDescription(clusterID, task)

		if g.Match(strings.ToLower(*taskDescription.Containers[0].Name)) {
			newTaskWrapper := TaskWrapper{
				HostPort:          *taskDescription.Containers[0].NetworkBindings[0].HostPort,
				Container:         strings.Split(*taskDescription.Containers[0].ContainerArn, "/")[1],
				ContainerInstance: strings.Split(*taskDescription.ContainerInstanceArn, "/")[1],
				ServiceName:       strings.ToLower(*taskDescription.Containers[0].Name),
				Task:              task,
			}

			slice = append(slice, newTaskWrapper)
		}
	}

	return slice, nil
}

func fetchContainerInstance(clusterID string, task TaskWrapper) (string, error) {

	params := &ecs.DescribeContainerInstancesInput{
		ContainerInstances: []*string{ // Required
			aws.String(task.ContainerInstance), // Required
			// More values...
		},
		Cluster: aws.String(clusterID),
	}

	resp, err := ecscli.DescribeContainerInstances(params)

	if err != nil {
		return "", err
	}

	return *resp.ContainerInstances[0].Ec2InstanceId, nil
}

func fetchEC2Instance(instanceID string) (EC2Wrapper, error) {

	params := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceID), // Required
			// More values...
		},
	}

	resp, err := ec2cli.DescribeInstances(params)

	if err != nil {
		return EC2Wrapper{}, err
	}

	newEC2Wrapper := EC2Wrapper{
		PrivateDNSName: *resp.Reservations[0].Instances[0].PrivateDnsName,
		PrivateIP:      *resp.Reservations[0].Instances[0].PrivateIpAddress,
		PublicDNSName:  *resp.Reservations[0].Instances[0].PublicDnsName,
		PublicIP:       *resp.Reservations[0].Instances[0].PublicIpAddress,
	}

	return newEC2Wrapper, nil
}

func makeMapping(clusterID string, task TaskWrapper, queue chan mapHelper, wg *sync.WaitGroup) {
	// Tell go routine has finish after function execution
	defer wg.Done()

	containerEC2InstanceID, _ := fetchContainerInstance(clusterID, task)
	ec2Instance, _ := fetchEC2Instance(containerEC2InstanceID)

	queue <- mapHelper{task.ServiceName, Microservice{
		Ec2Infos: ec2Instance,
		Task:     task,
	}}

}

func getMicroservices(clusterID string, tasks []TaskWrapper) (map[string][]Microservice, error) {

	// wait group for the goroutines we fire
	var wg sync.WaitGroup
	wg.Add(len(tasks))

	// channel fo goroutines to submit their work result
	queue := make(chan mapHelper)

	for _, task := range tasks {
		// fire goroutines
		go makeMapping(clusterID, task, queue, &wg)
	}

	// wait for the goroutines to finish and close the channel
	go func() {
		wg.Wait()
		close(queue)
	}()

	// final map for microservices
	mMap := make(map[string][]Microservice)

	for mHelper := range queue {
		mMap[mHelper.serviceName] = append(mMap[mHelper.serviceName], mHelper.microservice)
	}
	return mMap, nil
}

// GetServicesEndpoints returns ecs containers endpoints
func GetServicesEndpoints(clusterID string, region string, filter string) (map[string][]Microservice, error) {
	// initialize connection to aws with region config
	ecscli = ecscli.Initialize(region)
	ec2cli = ec2cli.Initialize(region)

	// Retrive all the tasks
	tasksIDs, err := fetchTasksIDs(clusterID)

	if err != nil {
		return nil, err
	}

	//Filter for tasks matching our serviceID
	tasks, err := filterTasks(clusterID, tasksIDs, filter)

	if err != nil {
		return nil, err
	}

	microservicesMap, err := getMicroservices(clusterID, tasks)

	if err != nil {
		return nil, err
	}

	return microservicesMap, nil
}
