package main

import (
	"flag"
	"fmt"
	"gitlab.com/scoro/infrastructure/mariabackup/Manager"
	"log"
	"os"
)

//backup command
var BackupCommand = flag.NewFlagSet("backup", flag.ExitOnError)
var Host = BackupCommand.String("host", "localhost", "database host")
var Port = BackupCommand.Int("port", 3306, "database port")
var Username = BackupCommand.String("username", "mariabackup", "database username")
var Password = BackupCommand.String("password", "", "database password")
var Type = BackupCommand.String("type", "full", "backup type - full|incremental")
var TargetDir = BackupCommand.String("target-dir", "/backup/mariabackup", "directory in which the backups will be placed")
var DataDir = BackupCommand.String("datadir", "/var/lib/mysql", "directory where the MySQL data is stored")

//restore command
var RestoreCommand = flag.NewFlagSet("restore", flag.ExitOnError)
var SourceDir = RestoreCommand.String("source-dir", "/backup/mariabackup", "directory in which the backups are stored")

func main() {

	log.SetFlags(log.Ldate | log.Ltime)

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

		log.Println("Backup data directory:", *DataDir)
		log.Println("Backup target directory:", *TargetDir)

		backup, err := Manager.CreateBackupManager(
			*TargetDir,
			*Host,
			*Port,
			*Username,
			*Password,
			*Type,
			*DataDir,
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

		log.Println("Todo: needs implementation..")
		log.Println("Source dir", *SourceDir)
	default:
		fmt.Printf("%q is not valid command\n", os.Args[1])
		return
	}
}
