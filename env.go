package chagent

import "os"

type NetflixEnv struct {
	Account     string
	AccountType string
	Environment string
	Region      string
	Stack       string
	Asg         string
	InstanceId  string
}

func NewNetflixEnv() *NetflixEnv {
	if IsLocalEnvironment() {
		logger.Infof(
			"Creating local environment since we did not find the NETFLIX_INSTANCE_ID environment variable")
		return &NetflixEnv{
			Account:     "ieptest",
			AccountType: "iep",
			Environment: "test",
			Region:      "us-east-1",
			Stack:       "dev",
			Asg:         "clickhousekeeper-newdev-v001",
			InstanceId:  "i-09380206b9f126aa3",
		}
	} else {
		stack := os.Getenv("NETFLIX_STACK")
		if stack == "" || stack == "newdev" {
			stack = "dev"
		}
		return &NetflixEnv{
			Account:     os.Getenv("NETFLIX_ACCOUNT"),
			AccountType: os.Getenv("NETFLIX_ACCOUNT_TYPE"),
			Environment: os.Getenv("NETFLIX_ENVIRONMENT"),
			Region:      os.Getenv("NETFLIX_REGION"),
			Stack:       stack,
			Asg:         os.Getenv("NETFLIX_AUTO_SCALE_GROUP"),
			InstanceId:  os.Getenv("NETFLIX_INSTANCE_ID"),
		}
	}
}

func IsLocalEnvironment() bool {
	// if the environment variable NETFLIX_INSTANCE_ID is missing (or empty)
	// then we consider it a local environment
	return os.Getenv("NETFLIX_INSTANCE_ID") == ""
}
