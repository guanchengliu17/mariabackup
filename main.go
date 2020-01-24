package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gitlab.com/scoro/infrastructure/mariabackup/Manager"
	"log"
	"os"
)

func main() {

	log.SetFlags(log.Ldate | log.Ltime)

	type ConfigManager Manager.ConfigManager

	var config ConfigManager
	configData, err := Manager.LoadConfiguration()

	if err != nil {
		log.Println("Configuration loading failed:", err)
		return
	}

	err = json.Unmarshal([]byte(configData), &config)

	if err != nil {
		log.Println("Configuration parsing failed:", err)
		return
	}

	//backup command
	BackupCommand := flag.NewFlagSet("backup", flag.ExitOnError)
	BackupCommandMySQLHost := BackupCommand.String("host", config.Backup.MySQLHost, "database host")
	BackupCommandMySQLPort := BackupCommand.Int("port", config.Backup.MySQLPort, "database port")
	BackupCommandMySQLUsername := BackupCommand.String("username", config.Backup.MySQLUsername, "database username")
	BackupCommandMySQLPassword := BackupCommand.String("password", config.Backup.MySQLPassword, "database password")
	BackupCommandMode := BackupCommand.String("mode", config.Backup.Mode, "backup mode - full|incremental")
	BackupCommandTargetDirectory := BackupCommand.String("target-dir", config.Backup.TargetDirectory, "directory in which the backups will be placed")
	BackupCommandMySQLDataDirectory := BackupCommand.String("datadir", config.Backup.MySQLDataDirectory, "directory where the MySQL data is stored")
	BackupCommandMariabackupBinary := BackupCommand.String("mariabackup-binary", config.MariabackupBinary, "mariabackup binary")
	BackupCommandPositionFile := BackupCommand.String("backup-position-file", config.PositionFile, "file where backup position is stored")

	//restore command
	RestoreCommand := flag.NewFlagSet("restore", flag.ExitOnError)
	RestoreCommandSourceDirectory := RestoreCommand.String("source-dir", config.Restore.SourceDirectory, "directory in which the backups are stored")
	RestoreCommandTargetDirectory := RestoreCommand.String("target-dir", config.Restore.TargetDirectory, "directory where the MySQL data is stored")
	RestoreCommandWorkDirectory := RestoreCommand.String("work-dir", config.Restore.WorkDirectory, "directory where temporary data is stored during restore process")
	RestoreCommandMariabackupBinary := RestoreCommand.String("mariabackup-binary", config.MariabackupBinary, "mariabackup binary")
	RestoreCommandPositionFile := RestoreCommand.String("backup-position-file", config.PositionFile, "file where backup position is stored")
	RestoreCommandMbstreamBinary := RestoreCommand.String("mbstream-binary", config.MbstreamBinary, "mbstream binary")

	if len(os.Args) < 2 {
		log.Println("Invalid number of arguments. Usage: ./<binary> <command>")
		return
	}

	switch os.Args[1] {
	case "backup":
		err := BackupCommand.Parse(os.Args[2:])

		if err != nil {
			log.Println("Parsing backup command failed:", err)
			return
		}
		//do backup

		log.Println("Backup data directory:", *BackupCommandMySQLDataDirectory)
		log.Println("Backup target directory:", *BackupCommandTargetDirectory)
		log.Println("Mariabackup binary:", *BackupCommandMariabackupBinary)
		log.Println("Backup position file:", *BackupCommandPositionFile)

		backup, err := Manager.CreateBackupManager(
			*BackupCommandTargetDirectory,
			*BackupCommandMySQLHost,
			*BackupCommandMySQLPort,
			*BackupCommandMySQLUsername,
			*BackupCommandMySQLPassword,
			*BackupCommandMode,
			*BackupCommandMySQLDataDirectory,
			*BackupCommandMariabackupBinary,
			*BackupCommandPositionFile,
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

	case "restore":
		err := RestoreCommand.Parse(os.Args[2:])
		if err != nil {
			log.Println("Parsing restore command failed:", err)
			return
		}
		//do restore

		log.Println("Restore source directory:", *RestoreCommandSourceDirectory)
		log.Println("Restore target directory:", *RestoreCommandTargetDirectory)

		restore, err := Manager.CreateRestoreManager(
			*RestoreCommandSourceDirectory,
			*RestoreCommandTargetDirectory,
			*RestoreCommandWorkDirectory,
			*RestoreCommandMariabackupBinary,
			*RestoreCommandPositionFile,
			*RestoreCommandMbstreamBinary,
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
