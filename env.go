package chagent

import "os"

type NetflixEnv struct {
	Account     string
	AccountType string
	Region      string
	Asg         string
	InstanceId  string
}

func NewNetflixEnv() *NetflixEnv {
	if isLocalEnv() {
		logger.Infof(
			"Creating local environment since we did not find the NETFLIX_INSTANCE_ID environment variable")
		return &NetflixEnv{
			Account:     "ieptest",
			AccountType: "iep",
			Region:      "us-east-1",
			Asg:         "clickhousekeeper-newdev-v001",
			InstanceId:  "i-09380206b9f126aa3",
		}
	} else {
		return &NetflixEnv{
			Account:     os.Getenv("NETFLIX_ACCOUNT"),
			AccountType: os.Getenv("NETFLIX_ACCOUNT_TYPE"),
			Region:      os.Getenv("NETFLIX_REGION"),
			Asg:         os.Getenv("NETFLIX_AUTO_SCALE_GROUP"),
			InstanceId:  os.Getenv("NETFLIX_INSTANCE_ID"),
		}
	}
}

func isLocalEnv() bool {
	// if the environment variable NETFLIX_INSTANCE_ID is missing (or empty)
	// then we consider it a local environment
	return os.Getenv("NETFLIX_INSTANCE_ID") == ""
}
