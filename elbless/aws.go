package elbless

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
)

// AWSECSClient defines interface for aws ecs calls
type AWSECSClient struct {
	conn *ecs.ECS
}

// Initialize the connection with specified region
func (c *AWSECSClient) Initialize(region string) AWSECSClient {
	var sess = session.New()

	return AWSECSClient{
		conn: ecs.New(sess, &aws.Config{Region: aws.String(region)}),
	}
}

//DescribeContainerInstances aws ecs passthrough
func (c *AWSECSClient) DescribeContainerInstances(input *ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error) {
	return c.conn.DescribeContainerInstances(input)
}

// DescribeTasks aws ecs passthrough
func (c *AWSECSClient) DescribeTasks(input *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error) {
	return c.conn.DescribeTasks(input)
}

// ListClusters aws ecs passthrough
func (c *AWSECSClient) ListClusters(input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error) {
	return c.conn.ListClusters(input)
}

//ListTasks aws ecs passthrough
func (c *AWSECSClient) ListTasks(input *ecs.ListTasksInput) (*ecs.ListTasksOutput, error) {
	return c.conn.ListTasks(input)
}

//AWSEC2Client defines interface for aws ecs calls
type AWSEC2Client struct {
	conn *ec2.EC2
}

// Initialize the connection with specified region
func (c *AWSEC2Client) Initialize(region string) AWSEC2Client {
	var sess = session.New()

	return AWSEC2Client{
		conn: ec2.New(sess, &aws.Config{Region: aws.String(region)}),
	}
}

//DescribeInstances aws ec2 passthrough
func (c *AWSEC2Client) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return c.conn.DescribeInstances(input)
}
