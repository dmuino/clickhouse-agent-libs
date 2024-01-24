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
	instanceInfo InstanceInfo
	asgName      string
	env          *NetflixEnv
	logger       *Logger
}

func NewSlotInfo(env *NetflixEnv) *SlotInfo {
	instanceInfo := InstanceInfo{
		Slot: -1,
	}
	return &SlotInfo{instanceInfo: instanceInfo, env: env, logger: GetLogger("Slotting")}
}

func (s *SlotInfo) SetLevel(level Level) {
	s.logger.level = level
}

func (s *SlotInfo) getAsgInfo(env *NetflixEnv, asg string) AsgInfo {
	baseUrl := getBaseUrl(env)
	url := fmt.Sprintf("%s/api/v1/autoScalingGroups/%s", baseUrl, asg)
	s.logger.Debugf("Getting all nodes for asg={} using url: %s", asg, url)

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

func (s *SlotInfo) slot() int {
	return s.instanceInfo.Slot
}

func (s *SlotInfo) GetMyInfo(env *NetflixEnv) InstanceInfo {
	if s.slot() == -1 {
		const maxRetries = 10
		for retry := 0; s.slot() == -1 && retry < maxRetries; retry++ {
			if retry > 0 {
				secondsToSleep := time.Duration(max(5*retry, 30)) * time.Second
				s.logger.Infof("Sleeping %v before retrying to get slot. Attempt %d of %d", secondsToSleep, retry, maxRetries)
				time.Sleep(secondsToSleep)
			}
			asgInfo := s.getAsgInfo(env, env.Asg)
			for _, instance := range asgInfo.Instances {
				if instance.InstanceId == env.InstanceId {
					s.instanceInfo = instance
					return s.instanceInfo
				}
			}
		}
	}
	return s.instanceInfo
}

func (s *SlotInfo) GetSlot(env *NetflixEnv) int {
	return s.GetMyInfo(env).Slot
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

func (s *SlotInfo) getAllAsgsInCluster(env *NetflixEnv, cluster string) []AsgInfo {
	baseUrl := getBaseUrl(env)
	url := fmt.Sprintf("%s/api/v1/clusters/%s?verbose=true", baseUrl, cluster)
	s.logger.Debugf("Getting all asgs from cluster using url: %s", url)
	// make http get request to get all nodes in our ASG from the slotting service
	// and return the list of nodes
	resp, err := http.Get(url)
	if err != nil {
		s.logger.Errorf("Error getting all nodes from cluster: %v", err)
		return []AsgInfo{}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Errorf("Error getting all nodes from cluster: %v", err)
		return []AsgInfo{}
	}

	// parse body as []AsgInfo
	var asgsCluster []AsgInfo
	err = json.Unmarshal(body, &asgsCluster)
	if err != nil {
		s.logger.Errorf("Error parsing output from slotting getting all nodes from cluster: %v", err)
		return []AsgInfo{}
	}

	return asgsCluster
}

func (s *SlotInfo) GetAllAsgsInCluster(env *NetflixEnv, cluster string) []AsgInfo {
	const sleepTime = 30
	const maxRetries = 15

	retry := 0
	for {
		asgs := s.getAllAsgsInCluster(env, cluster)
		if len(asgs) > 0 {
			return asgs
		}
		if retry < maxRetries {
			retry++
			s.logger.Infof("Waiting until the slotting service updates the list of ASGs for our cluster. Retrying in %d seconds", sleepTime)
			time.Sleep(sleepTime * time.Second)
		} else {
			s.logger.Fatalf("Could not find any ASGs in cluster %s", cluster)
		}
	}
}

// this is used for DNS purposes, so we want to return even instances that are out of service
// or in the previous ASG
func (s *SlotInfo) getAllNodesInCluster(env *NetflixEnv, cluster string) []InstanceInfo {
	asgsCluster := s.getAllAsgsInCluster(env, cluster)

	var instances []InstanceInfo
	for _, asg := range asgsCluster {
		// append all instances in asg to instances
		instances = append(instances, asg.Instances...)
	}
	return instances
}

func (s *SlotInfo) GetAllNodesInCluster(env *NetflixEnv, cluster string) []InstanceInfo {
	const sleepTime = 30
	const maxRetries = 15

	retry := 0
	for {
		nodes := s.getAllNodesInCluster(env, cluster)
		if len(nodes) > 0 && instanceIdInList(s.env.InstanceId, nodes) {
			return nodes
		}
		if retry < maxRetries {
			retry++
			s.logger.Infof("Waiting until the slotting service assigns a slot to this node. Retrying in %d seconds", sleepTime)
			time.Sleep(sleepTime * time.Second)
		} else {
			s.logger.Errorf("Could not find my instanceId: %s in the list of nodes: %v", s.env.InstanceId, nodes)
			return nodes
		}
	}
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
