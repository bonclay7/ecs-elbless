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

		newTaskWrapper := TaskWrapper{
			HostPort:          *taskDescription.Containers[0].NetworkBindings[0].HostPort,
			Container:         strings.Split(*taskDescription.Containers[0].ContainerArn, "/")[1],
			ContainerInstance: strings.Split(*taskDescription.ContainerInstanceArn, "/")[1],
			ServiceName:       strings.ToLower(*taskDescription.Containers[0].Name),
			Task:              task,
		}

		slice = append(slice, newTaskWrapper)
	}

	return slice
}

func fetchContainerInstance(clusterID string, task TaskWrapper, region string) string {
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

	newEC2Wrapper := EC2Wrapper{
		PrivateDNSName: *resp.Reservations[0].Instances[0].PrivateDnsName,
		PrivateIP:      *resp.Reservations[0].Instances[0].PrivateIpAddress,
		PublicDNSName:  *resp.Reservations[0].Instances[0].PublicDnsName,
		PublicIP:       *resp.Reservations[0].Instances[0].PublicIpAddress,
	}

	return newEC2Wrapper
}

func getMicroservices(clusterID string, tasks []TaskWrapper, region string) (MicroservicesMap map[string][]Microservice) {

	MicroservicesMap = make(map[string][]Microservice)

	for _, task := range tasks {
		containerEC2InstanceID := fetchContainerInstance(clusterID, task, region)
		ec2Instance := fetchEC2Instance(containerEC2InstanceID, region)

		newMicroservice := Microservice{
			Ec2Infos: ec2Instance,
			Task:     task,
		}

		MicroservicesMap[task.ServiceName] = append(MicroservicesMap[task.ServiceName], newMicroservice)
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
