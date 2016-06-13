package elbless

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
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

var sess = session.New()

// Fetch from AWS tasks created for an ECS cluster
func fetchTasksIDs(clusterID string, region string) ([]string, error) {
	svc := ecs.New(sess, &aws.Config{Region: aws.String(region)})

	params := &ecs.ListTasksInput{
		Cluster: aws.String(clusterID),
	}

	resp, err := svc.ListTasks(params)

	if err != nil {
		return nil, err
	}

	tasksSlice := make([]string, len(resp.TaskArns))

	for idx, task := range resp.TaskArns {
		tasksSlice[idx] = strings.Split(*task, "/")[1]
	}

	return tasksSlice, nil
}

func fetchTaskDescription(clusterID string, taskID string, region string) (*ecs.Task, error) {
	svc := ecs.New(sess, &aws.Config{Region: aws.String(region)})

	params := &ecs.DescribeTasksInput{
		Tasks: []*string{
			aws.String(taskID),
		},
		Cluster: aws.String(clusterID),
	}

	resp, err := svc.DescribeTasks(params)

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

func filterTasks(clusterID string, tasks []string, region string, filter string) ([]TaskWrapper, error) {

	var g glob.Glob

	slice := make([]TaskWrapper, 0, len(tasks))
	g = glob.MustCompile(strings.ToLower(filter))

	for _, task := range tasks {
		taskDescription, _ := fetchTaskDescription(clusterID, task, region)

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

func fetchContainerInstance(clusterID string, task TaskWrapper, region string) (string, error) {
	svc := ecs.New(sess, &aws.Config{Region: aws.String(region)})

	params := &ecs.DescribeContainerInstancesInput{
		ContainerInstances: []*string{ // Required
			aws.String(task.ContainerInstance), // Required
			// More values...
		},
		Cluster: aws.String(clusterID),
	}

	resp, err := svc.DescribeContainerInstances(params)

	if err != nil {
		return "", err
	}

	return *resp.ContainerInstances[0].Ec2InstanceId, nil
}

func fetchEC2Instance(instanceID string, region string) (EC2Wrapper, error) {
	svc := ec2.New(sess, &aws.Config{Region: aws.String(region)})

	params := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceID), // Required
			// More values...
		},
	}

	resp, err := svc.DescribeInstances(params)

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

func getMicroservices(clusterID string, tasks []TaskWrapper, region string) (map[string][]Microservice, error) {

	microservicesMap := make(map[string][]Microservice)

	for _, task := range tasks {
		containerEC2InstanceID, err := fetchContainerInstance(clusterID, task, region)

		if err != nil {
			return nil, err
		}

		ec2Instance, err := fetchEC2Instance(containerEC2InstanceID, region)

		if err != nil {
			return nil, err
		}

		newMicroservice := Microservice{
			Ec2Infos: ec2Instance,
			Task:     task,
		}
		microservicesMap[task.ServiceName] = append(microservicesMap[task.ServiceName], newMicroservice)
	}

	return microservicesMap, nil
}

// GetServicesEndpoints returns ecs containers endpoints
func GetServicesEndpoints(clusterID string, region string, filter string) (map[string][]Microservice, error) {

	// Retrive all the tasks
	tasksIDs, err := fetchTasksIDs(clusterID, region)

	if err != nil {
		return nil, err
	}

	//Filter for tasks matching our serviceID
	tasks, err := filterTasks(clusterID, tasksIDs, region, filter)

	if err != nil {
		return nil, err
	}

	microservicesMap, err := getMicroservices(clusterID, tasks, region)

	if err != nil {
		return nil, err
	}

	return microservicesMap, nil
}
