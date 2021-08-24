package Manager

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	MariaBackupBinary string  `json:"maria_backup_binary"`
	MbStreamBinary    string  `json:"mb_stream_binary"`
	PositionFile      string  `json:"position_file"`
	Backup            backup  `json:"backup"`
	Restore           restore `json:"restore"`
	ParallelThreads   int     `json:"parallel_threads"`
	GzipThreads       int     `json:"compression_threads"`
	GzipBlockSize     int     `json:"compression_block_size"`
}

type restore struct {
	SourceDirectory string `json:"source_directory"`
	TargetDirectory string `json:"target_directory"`
	WorkDirectory   string `json:"work_directory"`
}

type backup struct {
	TargetDirectory string `json:"target_directory"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	Mode            string `json:"mode"`
	DataDirectory   string `json:"data_directory"`
	S3              struct {
		Region    string `json:"region"`
		AccessKey string `json:"access_key"`
		Secret    string `json:"secret"`
		Bucket    string `json:"bucket"`
	}
}

func CreateNewConfig() *Config {

	config := &Config{Restore: restore{
		SourceDirectory: "/backup/mariabackup",
		TargetDirectory: "/var/lib/mysql",
		WorkDirectory:   "/backup/mariabackup/restore",
	},
		Backup: backup{
			TargetDirectory: "/backup/mariabackup",
			Host:            "localhost",
			Port:            3306,
			Username:        "mariabackup",
			Password:        "",
			Mode:            "full",
			DataDirectory:   "/var/lib/mysql",
		},
		MariaBackupBinary: "/usr/bin/mariabackup",
		MbStreamBinary:    "/usr/bin/mbstream",
		PositionFile:      "/backup/mariabackup/mariabackup.pos",
		GzipBlockSize:     512 << 10,
		GzipThreads:       8,
		ParallelThreads:   4,
	}

	return config
}

func (c *Config) Load(file string) error {
	data, err := ioutil.ReadFile(file)

	if err != nil {
		return err
	}

	return json.Unmarshal(data, &c)
}

func (c *Config) CheckIfExists(file string) error {
	_, err := os.Stat(file)
	return err
}

func (c *Config) Save(file string) error {
	payload, err := json.MarshalIndent(&c, "", "\t")

	if err != nil {
		return err
	}

	fh, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	defer fh.Close()

	_, err = fh.Write(payload)

	return err
}
