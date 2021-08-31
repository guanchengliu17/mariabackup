package main

import (
	"flag"
	"fmt"
	"github.com/karlmjogila/mariabackup/Manager"
	"log"
	"os"
	"path/filepath"
)

//backup command
var Backup = flag.NewFlagSet("backup", flag.ExitOnError)
var BackupHost = Backup.String("host", "", "database host")
var BackupPort = Backup.Int("port", 0, "database port")
var BackupUsername = Backup.String("username", "", "database username")
var BackupPassword = Backup.String("password", "", "database password")
var BackupMode = Backup.String("mode", "", "backup mode - full|incremental")
var BackupTargetDirectory = Backup.String("target-dir", "", "directory in which the backups will be placed")
var BackupDataDirectory = Backup.String("datadir", "", "directory where the MySQL data is stored")
var BackupMariaBackupBinary = Backup.String("mariabackup-binary", "", "mariabackup binary")
var BackupPositionFile = Backup.String("backup-position-file", "", "file where backup position is stored")
var BackupConfigFile = Backup.String("config-file", "", "configuration file")
var BackupParallelThreads = Backup.Int("parallel-threads", 0, "parallel threads for mariabackup")
var BackupGzipThreads = Backup.Int("gzip-threads", 0, "gzip number of threads")
var BackupGzipBlockSize = Backup.Int("gzip-block", 0, "number of bytes gzip processes per cycle")
var BackupToS3 = Backup.Bool("backup-to-s3", false, "When true upload to S3")

//restore command
var Restore = flag.NewFlagSet("restore", flag.ExitOnError)
var RestoreSourceDirectory = Restore.String("source-dir", "", "directory in which the backups are stored")
var RestoreTargetDirectory = Restore.String("target-dir", "", "directory where the MySQL data is stored")
var RestoreWorkDirectory = Restore.String("work-dir", "", "directory where temporary data is stored during restore process")
var RestoreMariaBackupBinary = Restore.String("mariabackup-binary", "", "mariabackup binary")
var RestorePositionFile = Restore.String("backup-position-file", "", "file where backup position is stored")
var RestoreMbStreamBinary = Restore.String("mbstream-binary", "", "mbstream binary")
var RestoreConfigFile = Restore.String("config-file", "", "configuration file")
var RestoreGzipThreads = Restore.Int("gzip-threads", 0, "gzip number of threads")
var RestoreGzipBlockSize = Restore.Int("gzip-block", 0, "number of bytes gzip processes per cycle")
var RestoreFromS3 = Restore.Bool("restore-from-s3", false, "When true restore from S3")
var RestoreDate = Restore.String("restore-date", "", "backup creation date from S3, format YYYY-MM-DD")

func main() {
	log.SetFlags(log.Ldate | log.Ltime)

	if len(os.Args) < 2 {
		log.Println("Invalid number of arguments. Usage: " + os.Args[0] + " <command>")
		return
	}

	switch os.Args[1] {
	case "backup":
		err := Backup.Parse(os.Args[2:])

		if err != nil {
			log.Println("Parsing backup command failed:", err)
			return
		}
		//do backup

		config := loadConfig()

		log.Println("Backup data directory:", config.Backup.DataDirectory)
		log.Println("Backup target directory:", config.Backup.TargetDirectory)
		log.Println("MariaBackup binary:", config.MariaBackupBinary)
		log.Println("Backup position file:", config.PositionFile)

		backup, err := Manager.CreateBackupManager(
			config.Backup.TargetDirectory,
			config.Backup.Host,
			config.Backup.Port,
			config.Backup.Username,
			config.Backup.Password,
			config.Backup.Mode,
			config.Backup.DataDirectory,
			config.MariaBackupBinary,
			config.PositionFile,
			config.GzipBlockSize,
			config.GzipThreads,
			config.ParallelThreads,
		)

		if err != nil {
			log.Printf("Failed to initialize backup")
			return
		}

		err = backup.Backup()

		if err != nil {
			log.Println("Backup has failed:", err)
			return
		}

		log.Printf("Backup successfully finished")

		if *BackupToS3 {

			encrypt := Manager.Encrypt{}

			err1 := encrypt.Encrypt(
				filepath.Join(config.S3.UploadDirectory, "backup.gz"),
				filepath.Join(config.S3.UploadDirectory, "backup.gz.enc"),
				"enc_key",
				1024)

			if err1 != nil {
				return
			}

			if err != nil {
				log.Println("Failed to initialize encryption")

			}

			upload, err := Manager.CreateS3Manager(
				config.S3.AccessKey,
				config.S3.Region,
				config.S3.Bucket,
				config.S3.Secret,
			)

			if err != nil {
				fmt.Println(err)
			}

			upload.Upload(config.S3.UploadDirectory)
		}

	case "restore":
		err := Restore.Parse(os.Args[2:])
		if err != nil {
			log.Println("Parsing restore command failed:", err)
			return
		}
		//do restore

		config := loadConfig()

		log.Println("Restore source directory:", config.Restore.SourceDirectory)
		log.Println("Restore target directory:", config.Restore.TargetDirectory)

		if *RestoreFromS3 {
			err := Restore.Parse(os.Args[4:])
			if err != nil {
				log.Println("Parsing restore command failed:", err)
				return
			}
			download, err := Manager.CreateS3Manager(
				config.S3.AccessKey,
				config.S3.Region,
				config.S3.Bucket,
				config.S3.Secret,
			)

			if err != nil {
				fmt.Println(err)
			}
			download.Download(config.S3.UploadDirectory, *RestoreDate)

			decrypt := Manager.Decrypt{}

			err1 := decrypt.Decrypt(
				filepath.Join(config.S3.UploadDirectory, "backup.gz.enc"),
				filepath.Join(config.S3.UploadDirectory, "backup.gz"),
				"enc_key",
				1024)
			if err1 != nil {
				return
			}

			if err != nil {
				log.Println("Failed to initialize decryption")

			}
		}

		restore, err := Manager.CreateRestoreManager(
			config.Restore.SourceDirectory,
			config.Restore.TargetDirectory,
			config.Restore.WorkDirectory,
			config.MariaBackupBinary,
			config.PositionFile,
			config.MbStreamBinary,
			config.GzipBlockSize,
			config.GzipThreads,
		)

		if err != nil {
			log.Printf("Failed to initialize restore")
			return
		}

		err = restore.Restore()

		if err != nil {
			log.Println("Restore has failed:", err)
			return
		}

		log.Printf("Restore successfully finished")

	default:
		fmt.Printf("%q is not valid command\n", os.Args[1])
		return
	}
}

func loadConfig() *Manager.Config {
	config := Manager.CreateNewConfig()

	configFile := "config.json" //default

	//determine the config file based on the command

	if Backup.Parsed() {
		if len(*BackupConfigFile) > 0 {
			configFile = *BackupConfigFile
		}
	}

	if Restore.Parsed() {
		if len(*RestoreConfigFile) > 0 {
			configFile = *RestoreConfigFile
		}
	}

	if config.CheckIfExists(configFile) != nil {
		err := config.Save(configFile) //try to create config file
		if err != nil {
			log.Fatalln("Failed to create config file", err)
		}
	}
	err := config.Load("config.json")

	if err != nil {
		log.Fatalln("Failed to read config file: ", err)
	}

	if Backup.Parsed() {

		if len(*BackupHost) > 0 {
			config.Backup.Host = *BackupHost
		}

		if *BackupPort > 0 {
			config.Backup.Port = *BackupPort
		}

		if len(*BackupUsername) > 0 {
			config.Backup.Username = *BackupUsername
		}

		if len(*BackupPassword) > 0 {
			config.Backup.Password = *BackupPassword
		}

		if len(*BackupMode) > 0 {
			config.Backup.Mode = *BackupMode
		}

		if len(*BackupTargetDirectory) > 0 {
			config.Backup.TargetDirectory = *BackupTargetDirectory
		}

		if len(*BackupDataDirectory) > 0 {
			config.Backup.DataDirectory = *BackupDataDirectory
		}

		if len(*BackupMariaBackupBinary) > 0 {
			config.MariaBackupBinary = *BackupMariaBackupBinary
		}

		if len(*BackupPositionFile) > 0 {
			config.PositionFile = *BackupPositionFile
		}

		if *BackupParallelThreads > 0 {
			config.ParallelThreads = *BackupParallelThreads
		}

		if *BackupGzipThreads > 0 {
			config.GzipThreads = *BackupGzipThreads
		}

		if *BackupGzipBlockSize > 0 {
			config.GzipBlockSize = *BackupGzipBlockSize
		}

	}

	if Restore.Parsed() {

		if len(*RestoreSourceDirectory) > 0 {
			config.Restore.SourceDirectory = *RestoreSourceDirectory
		}

		if len(*RestoreTargetDirectory) > 0 {
			config.Restore.TargetDirectory = *RestoreTargetDirectory
		}

		if len(*RestoreWorkDirectory) > 0 {
			config.Restore.WorkDirectory = *RestoreWorkDirectory
		}

		if len(*RestoreMariaBackupBinary) > 0 {
			config.MariaBackupBinary = *RestoreMariaBackupBinary
		}

		if len(*RestorePositionFile) > 0 {
			config.PositionFile = *RestorePositionFile
		}

		if len(*RestoreMbStreamBinary) > 0 {
			config.MbStreamBinary = *RestoreMbStreamBinary
		}

		if *RestoreGzipThreads > 0 {
			config.GzipThreads = *RestoreGzipThreads
		}

		if *RestoreGzipBlockSize > 0 {
			config.GzipBlockSize = *RestoreGzipBlockSize
		}
	}

	return config
}
