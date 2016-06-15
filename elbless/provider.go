package elbless

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
)

// AWSECS defines method (we use) for ecs sdk
type AWSECS interface {
	DescribeContainerInstances(*ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error)
	DescribeTasks(*ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error)
	Initialize(region string) AWSECSClient
	ListClusters(*ecs.ListClustersInput) (*ecs.ListClustersOutput, error)
	ListTasks(*ecs.ListTasksInput) (*ecs.ListTasksOutput, error)
}

// AWSEC2 defines method (we use) for ec2 sdk
type AWSEC2 interface {
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
	Initialize(region string) AWSECSClient
}
