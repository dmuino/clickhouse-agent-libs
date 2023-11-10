package chagent

import (
	"fmt"
	"testing"
)

var testEnv = &NetflixEnv{
	AccountType: "iep",
	Region:      "us-east-1",
	Account:     "ieptest",
	Asg:         "clickhousekeeper-newdev-v001",
	InstanceId:  "i-1",
}

func TestGetBaseUrl(t *testing.T) {
	actual := getBaseUrl(testEnv)
	expected := "https://slotting-iep.us-east-1.ieptest.netflix.net"
	if actual != expected {
		fmt.Printf("actual: %s != expected: %s\n", actual, expected)
		t.Fail()
	}
}
