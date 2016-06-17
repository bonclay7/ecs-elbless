package elbless

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecs"
)

type ECSMock struct {
}

const ballotX = "\u2717"
const checkMark = "\u2713"

func (c *ECSMock) ListTasks(input *ecs.ListTasksInput) (*ecs.ListTasksOutput, error) {

	var err error
	var output *ecs.ListTasksOutput

	switch *input.Cluster {
	case "cluster-do-not-exist":
		err = fmt.Errorf("i don't exist")
		output = nil

	case "valid-cluster":
		taskone := "arn:aws:ecs:eu-west-1:658452139221:task/166c6aa6-13d2-4f77-b176-3d5a33c1ae3a"

		tasksArn := []*string{
			&taskone,
		}

		output = &ecs.ListTasksOutput{
			NextToken: nil,
			TaskArns:  tasksArn,
		}
		err = nil
	}

	return output, err
}

// Initialize the connection with specified region
func (c *ECSMock) Initialize(region string) AWSECSClient {
	return AWSECSClient{nil}
}

//DescribeContainerInstances aws ecs passthrough
func (c *ECSMock) DescribeContainerInstances(input *ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error) {
	return nil, nil
}

// DescribeTasks aws ecs passthrough
func (c *ECSMock) DescribeTasks(input *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error) {
	return nil, nil
}

// ListClusters aws ecs passthrough
func (c *ECSMock) ListClusters(input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error) {
	return nil, nil
}

func TestFetchTasksIDs(t *testing.T) {

	cases := map[string]struct {
		cluster  string
		expected []string
	}{
		"invalid-cluster": {"cluster-do-not-exist", nil},
		"valid-cluster":   {"valid-cluster", []string{"166c6aa6-13d2-4f77-b176-3d5a33c1ae3a"}},
	}

	cli := ECSMock{}

	for k, tc := range cases {
		t.Logf("Given case: %s", k)
		{
			resp, _ := fetchTasksIDs(tc.cluster, &cli)

			if reflect.DeepEqual(resp, tc.expected) {
				t.Logf("  Actual is %v, Got %v : %s", tc.expected, resp, checkMark)
			} else {
				t.Errorf("  Actual is %v, Got %v : %s", tc.expected, resp, ballotX)
			}

		}
	}

}
