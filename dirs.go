package chagent

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
)

func clickhouseDir(dir string, clickhouseUid, clickhouseGid int) {
	err := os.MkdirAll(dir, os.ModePerm)
	logger.CheckErr(err)
	err = os.Chown(dir, clickhouseUid, clickhouseGid)
	logger.CheckErr(err)
}

func SetupDirs(serviceName string) {
	// Get the user clickhouse which will own the directories
	clickhouseUser, err := user.Lookup("clickhouse")
	if err != nil {
		logger.Fatalf("Error looking up clickhouse user: %v", err)
	}

	clickhouseUid, err := strconv.Atoi(clickhouseUser.Uid)
	logger.CheckErr(err)
	clickhouseGid, err := strconv.Atoi(clickhouseUser.Gid)
	logger.CheckErr(err)

	// Remove the directory /var/log/$serviceName
	origLogDir := fmt.Sprintf("/var/log/%s", serviceName)
	if err := os.RemoveAll(origLogDir); err != nil {
		logger.Fatalf("Error removing %s: %v", origLogDir, err)
	}
	if err := os.RemoveAll("/var/lib/clickhouse"); err != nil {
		logger.Fatalf("Error removing /var/lib/clickhouse: %v", err)
	}

	varLibDir := fmt.Sprintf("/var/lib/%s", serviceName)
	if err := os.RemoveAll(varLibDir); err != nil {
		logger.Fatalf("Error removing %s: %v", varLibDir, err)
	}
	clickhouseDir(varLibDir, clickhouseUid, clickhouseGid)

	dataDir := fmt.Sprintf("/data/%s", serviceName)
	clickhouseDir(dataDir, clickhouseUid, clickhouseGid)
	// Create a symbolic link from /data/$serviceName to /var/lib/$serviceName
	if err := os.Symlink(dataDir, varLibDir); err != nil {
		logger.Fatalf("Error creating symbolic link: %v", err)
	}

	clickhouseDir(fmt.Sprintf("/logs/%s", serviceName), clickhouseUid, clickhouseGid)
	clickhouseDir("/run/"+serviceName, clickhouseUid, clickhouseGid)
	clickhouseDir("/data/clickhouse", clickhouseUid, clickhouseGid)

	// Create a symbolic link from /logs/$service_name to /var/log/$service_name
	if err := os.Symlink(fmt.Sprintf("/logs/%s", serviceName), origLogDir); err != nil {
		logger.Fatalf("Error creating symbolic link: %v", err)
	}

	const origDataDir = "/var/lib/clickhouse"
	_ = os.RemoveAll(origDataDir)
	const finalDataDir = "/data/clickhouse"
	clickhouseDir(finalDataDir, clickhouseUid, clickhouseGid)

	// Create a symbolic link from /data/clickhouse to /var/lib/clickhouse
	if err := os.Symlink(finalDataDir, "/var/lib/clickhouse"); err != nil {
		logger.Fatalf("Error creating symbolic link: %v", err)
	}

	logger.Infof("Directories setup successfully in order to run %s.", serviceName)
}
