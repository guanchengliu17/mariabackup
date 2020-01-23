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

	MariabackupBinary := flag.String("mariabackup-binary", config.MariabackupBinary, "mariabackup binary")

	//backup command
	BackupCommand := flag.NewFlagSet("backup", flag.ExitOnError)
	BackupCommandHost := BackupCommand.String("host", config.Backup.MySQLHost, "database host")
	BackupCommandPort := BackupCommand.Int("port", config.Backup.MySQLPort, "database port")
	BackupCommandUsername := BackupCommand.String("username", config.Backup.MySQLUsername, "database username")
	BackupCommandPassword := BackupCommand.String("password", config.Backup.MySQLPassword, "database password")
	BackupCommandMode := BackupCommand.String("mode", config.Backup.Mode, "backup mode - full|incremental")
	BackupCommandTargetDirectory := BackupCommand.String("target-dir", config.Backup.TargetDirectory, "directory in which the backups will be placed")
	BackupCommandDataDirectory := BackupCommand.String("datadir", config.Backup.MySQLDataDirectory, "directory where the MySQL data is stored")

	//restore command
	RestoreCommand := flag.NewFlagSet("restore", flag.ExitOnError)
	RestoreCommandSourceDirectory := RestoreCommand.String("source-dir", config.Restore.SourceDirectory, "directory in which the backups are stored")
	RestoreCommandTargetDirectory := RestoreCommand.String("target-dir", config.Restore.TargetDirectory, "directory where the MySQL data is stored")

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

		log.Println("Backup data directory:", *BackupCommandDataDirectory)
		log.Println("Backup target directory:", *BackupCommandTargetDirectory)
		log.Println("Mariabackup binary:", *MariabackupBinary)

		backup, err := Manager.CreateBackupManager(
			*BackupCommandTargetDirectory,
			*BackupCommandHost,
			*BackupCommandPort,
			*BackupCommandUsername,
			*BackupCommandPassword,
			*BackupCommandMode,
			*BackupCommandDataDirectory,
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

		restore, err := Manager.CreateRestoreManager(*RestoreCommandSourceDirectory, *RestoreCommandTargetDirectory)

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
