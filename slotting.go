package chagent

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func getBaseUrl(env *NetflixEnv) string {
	return fmt.Sprintf("https://slotting-%s.%s.%s.netflix.net", env.AccountType, env.Region, env.Account)
}

type InstanceInfo struct {
	InstanceId       string `json:"instanceId"`
	IPv6Address      string `json:"ipv6Address"`
	PrivateIpAddress string `json:"privateIpAddress"`
	Slot             int    `json:"slot"`
	LaunchTime       int64  `json:"launchTime"`
	ImageId          string `json:"imageId"`
	InstanceType     string `json:"instanceType"`
	AvailabilityZone string `json:"availabilityZone"`
	LifecycleState   string `json:"lifecycleState"`
}

type AsgInfo struct {
	Name            string         `json:"name"`
	Cluster         string         `json:"cluster"`
	CreatedTime     int64          `json:"createdTime"`
	DesiredCapacity int            `json:"desiredCapacity"`
	MaxSize         int            `json:"maxSize"`
	MinSize         int            `json:"minSize"`
	IsDisabled      bool           `json:"isDisabled"`
	Instances       []InstanceInfo `json:"instances"`
}

type SlotInfo struct {
	Slot   int `json:"slot"`
	env    *NetflixEnv
	logger *Logger
}

func NewSlotInfo(env *NetflixEnv) *SlotInfo {
	return &SlotInfo{Slot: -1, env: env, logger: GetLogger("Slotting")}
}

func (s *SlotInfo) GetSlot(env *NetflixEnv) int {
	if s.Slot == -1 {
		const maxRetries = 10
		for retry := 0; s.Slot == -1 && retry < maxRetries; retry++ {
			if retry > 0 {
				secondsToSleep := time.Duration(max(5*retry, 30)) * time.Second
				s.logger.Infof("Sleeping %v before retrying to get slot. Attempt %d of %d", secondsToSleep, retry, maxRetries)
				time.Sleep(secondsToSleep)
			}
			asgInfo := s.getAsgInfo(env, env.Asg)
			for _, instance := range asgInfo.Instances {
				if instance.InstanceId == env.InstanceId {
					s.Slot = instance.Slot
					return s.Slot
				}
			}
		}
	}
	return s.Slot
}

func (s *SlotInfo) getAsgInfo(env *NetflixEnv, asg string) AsgInfo {
	baseUrl := getBaseUrl(env)
	url := fmt.Sprintf("%s/api/v1/autoScalingGroups/%s", baseUrl, asg)
	s.logger.Infof("Getting all nodes from our ASG using url: %s", url)

	// make http get request to get all nodes in our ASG from the slotting service
	// and return the list of nodes
	resp, err := http.Get(url)
	s.logger.CheckErr(err)

	body, err := io.ReadAll(resp.Body)
	s.logger.CheckErr(err)

	// parse body as AsgInfo
	var asgInfo AsgInfo
	err = json.Unmarshal(body, &asgInfo)
	s.logger.CheckErr(err)

	return asgInfo
}

func (s *SlotInfo) GetAllNodes() []InstanceInfo {
	asgInfo := s.getAsgInfo(s.env, s.env.Asg)

	// return the list of instanceIds
	var instances []InstanceInfo
	for _, instance := range asgInfo.Instances {
		if instance.LifecycleState == "InService" || instance.LifecycleState == "Pending" {
			instances = append(instances, instance)
		} else {
			s.logger.Debugf("Skipping instance %s with lifecycleState %s", instance.InstanceId, instance.LifecycleState)
		}
	}
	return instances
}

func instanceIdInList(instanceId string, nodes []InstanceInfo) bool {
	for _, node := range nodes {
		if node.InstanceId == instanceId {
			return true
		}
	}
	return false
}

func (s *SlotInfo) GetAllNodesWithRetries() []InstanceInfo {
	const sleepTime = 30
	const maxRetries = 15

	retry := 0
	for {
		nodes := s.GetAllNodes()
		if len(nodes) > 0 && instanceIdInList(s.env.InstanceId, nodes) {
			return nodes
		}
		if retry < maxRetries {
			retry++
			s.logger.Infof("Waiting until the slotting service assigns a slot to this node. Retrying in %d seconds", sleepTime)
			time.Sleep(sleepTime * time.Second)
		} else {
			s.logger.Fatalf("Could not find my instanceId: %s in the list of nodes: %v", s.env.InstanceId, nodes)
		}
	}
}
