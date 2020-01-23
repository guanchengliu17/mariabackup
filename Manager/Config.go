package Manager

import (
	"encoding/json"
	"log"
)

type ConfigManager struct {
	MariabackupBinary string
	MbstreamBinary    string
	PositionFile      string
	Backup            struct {
		TargetDirectory    string
		MySQLHost          string
		MySQLPort          int
		MySQLUsername      string
		MySQLPassword      string
		Mode               string
		MySQLDataDirectory string
	}
	Restore struct {
		SourceDirectory string
		TargetDirectory string
		WorkDirectory   string
	}
}

func LoadConfiguration() ([]byte, error) {

	jsonData := &ConfigManager{
		MariabackupBinary: "/usr/bin/mariabackup",
		MbstreamBinary:    "/usr/bin/mbstream",
		PositionFile:      "/backup/mariabackup/mariabackup.pos",
	}

	jsonData.Backup.MySQLHost = "localhost"
	jsonData.Backup.MySQLPort = 3306
	jsonData.Backup.MySQLUsername = "mariabackup"
	jsonData.Backup.MySQLPassword = ""
	jsonData.Backup.Mode = "full"
	jsonData.Backup.TargetDirectory = "/backup/mariabackup"
	jsonData.Backup.MySQLDataDirectory = "/var/lib/mysql"

	jsonData.Restore.SourceDirectory = "/backup/mariabackup"
	jsonData.Restore.TargetDirectory = "/var/lib/mysql"
	jsonData.Restore.WorkDirectory = "/backup/mariabackup/restore"

	var config []byte
	config, err := json.Marshal(jsonData)

	if err != nil {
		log.Println(err)
	}

	return config, nil
}
