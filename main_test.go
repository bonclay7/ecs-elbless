package main

import "testing"

func TestFetchTaskIDWithInvalidClusterID(t *testing.T) {
	if fetchTasksIDs("i-do-not-exist") != nil {
		t.Error("Invalid clusterID should be null")
	}

}
