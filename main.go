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
var BackupHost = BackupCommand.String("host", "localhost", "database host")
var BackupPort = BackupCommand.Int("port", 3306, "database port")
var BackupUsername = BackupCommand.String("username", "mariabackup", "database username")
var BackupPassword = BackupCommand.String("password", "", "database password")
var BackupType = BackupCommand.String("type", "full", "backup type - full|incremental")
var BackupTargetDir = BackupCommand.String("target-dir", "/backup/mariabackup", "directory in which the backups will be placed")
var BackupDataDir = BackupCommand.String("datadir", "/var/lib/mysql", "directory where the MySQL data is stored")

//restore command
var RestoreCommand = flag.NewFlagSet("restore", flag.ExitOnError)
var RestoreSourceDir = RestoreCommand.String("source-dir", "/backup/mariabackup", "directory in which the backups are stored")
var RestoreTargetDir = RestoreCommand.String("target-dir", "/var/lib/mysql", "directory where the MySQL data is stored")

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

		log.Println("Backup data directory:", *BackupDataDir)
		log.Println("Backup target directory:", *BackupTargetDir)

		backup, err := Manager.CreateBackupManager(
			*BackupTargetDir,
			*BackupHost,
			*BackupPort,
			*BackupUsername,
			*BackupPassword,
			*BackupType,
			*BackupDataDir,
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

		log.Println("Restore source directory:", *RestoreSourceDir)
		log.Println("Restore target directory:", *RestoreTargetDir)

		restore, err := Manager.CreateRestoreManager(*RestoreSourceDir, *RestoreTargetDir)

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
