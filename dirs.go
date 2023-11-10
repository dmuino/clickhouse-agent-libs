package chagent

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
)

func SetupDirs(serviceName string) {
	// Remove the directory /var/log/$serviceName
	origLogDir := fmt.Sprintf("/var/log/%s", serviceName)
	if err := os.RemoveAll(origLogDir); err != nil {
		logger.Fatalf("Error removing %s: %v", origLogDir, err)
	}
	if err := os.RemoveAll("/var/lib/clickhouse"); err != nil {
		logger.Fatalf("Error removing /var/lib/clickhouse: %v", err)
	}

	finalLogDir := fmt.Sprintf("/logs/%s", serviceName)
	err := os.MkdirAll(finalLogDir, os.ModePerm)
	logger.CheckErr(err)

	clickhouseUser, err := user.Lookup("clickhouse")
	logger.CheckErr(err)

	err = os.MkdirAll("/run/"+serviceName, os.ModePerm)
	logger.CheckErr(err)
	clickhouseUid, err := strconv.Atoi(clickhouseUser.Uid)
	logger.CheckErr(err)
	clickhouseGid, err := strconv.Atoi(clickhouseUser.Gid)
	logger.CheckErr(err)
	err = os.Chown(finalLogDir, clickhouseUid, clickhouseGid)
	logger.CheckErr(err)
	err = os.Chown("/run/"+serviceName, clickhouseUid, clickhouseGid)
	logger.CheckErr(err)
	err = os.MkdirAll("/data/clickhouse", os.ModePerm)
	logger.CheckErr(err)
	err = os.Chown("/data/clickhouse", clickhouseUid, clickhouseGid)
	logger.CheckErr(err)

	// Create a symbolic link from /logs/clickhouse-server to /var/log/clickhouse-server
	if err := os.Symlink("/logs/clickhouse-server", "/var/log/clickhouse-server"); err != nil {
		logger.Fatalf("Error creating symbolic link: %v", err)
	}

	const origDataDir = "/var/lib/clickhouse"
	_ = os.RemoveAll(origDataDir)
	const finalDataDir = "/data/clickhouse"
	if err := os.MkdirAll(finalDataDir, os.ModePerm); err != nil {
		logger.Fatalf("Error creating %s directory: %v", finalDataDir, err)
	}

	// Create a symbolic link from /data/clickhouse to /var/lib/clickhouse
	if err := os.Symlink("/data/clickhouse", "/var/lib/clickhouse"); err != nil {
		logger.Fatalf("Error creating symbolic link: %v", err)
	}

	logger.Infof("Directories setup successfully in order to run %s.", serviceName)
}
