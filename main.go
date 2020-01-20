package main

import (
	"compress/gzip"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

var BackupCommand = flag.NewFlagSet("backup", flag.ExitOnError)
var Host = BackupCommand.String("host", "localhost", "database host")
var Port = BackupCommand.Int("port", 3306, "database port")
var Username = BackupCommand.String("username", "mariabackup", "database username")
var Password = BackupCommand.String("password", "", "database password")
var Type = BackupCommand.String("type", "full", "backup type")
var TargetDir = BackupCommand.String("target_dir", "/backup/mariabackup", "directory in which the backups will be placed")
var DataDir = BackupCommand.String("datadir", "/var/lib/mysql", "directory where the MySQL data is stored")

func main() {

	log.SetFlags(log.Ldate | log.Ltime)

	if len(os.Args) < 2 {
		log.Println("Invalid number of arguments. Usage: backup <command>")
		return
	}

	switch os.Args[1] {
	case "backup":
		err := BackupCommand.Parse(os.Args)

		if err != nil {
			log.Println("Parsing backup commands failed:", err)
			return
		}
	default:
		fmt.Printf("%q is not valid command.\n", os.Args[1])
		return
	}

	log.Println("Backup data directory:", *DataDir)
	log.Println("Backup target directory:", *TargetDir)

	err := backup(*TargetDir)

	if err != nil {
		log.Println("Backup aborted:", err)
	}

}

func backup(target string) error {

	backupPath := ""
	backupCmd := &exec.Cmd{}
	backupPos := 0
	incrementalBaseDir := ""

	if *Type == "full" {
		err := os.RemoveAll(*TargetDir)

		if err != nil {
			return errors.New(fmt.Sprintf("[Full backup]> Failed to remove targetDir, %v", err))
		}

		backupPath = filepath.Join(*TargetDir, "full")
		err = os.MkdirAll(backupPath, 750)

		if err != nil {
			return errors.New(fmt.Sprintf("[Full backup]> Making directories failed, %v", err))
		}

		backupCmd = exec.Command("mariabackup",
			"--host="+*Host,
			"--port="+strconv.Itoa(*Port),
			"--user="+*Username,
			"--password="+*Password,
			"--backup",
			"--version-check",
			"--datadir="+*DataDir,
			"--target_dir="+backupPath,
			"--extra-lsndir="+backupPath,
			"--stream=xbstream",
		)
		backupPos = 0
	} else if *Type == "incr" {

		data, err := ioutil.ReadFile(filepath.Join(*TargetDir, "mariabackup.pos"))

		if err != nil {
			return errors.New(fmt.Sprintf("[Incremental backup]> Failed to read backup position file, %v", err))
		}

		loadedPosition, err := strconv.Atoi(string(data))

		if loadedPosition == 0 {
			incrementalBaseDir = filepath.Join(*TargetDir, "full")
		} else {
			incrementalBaseDir = filepath.Join(*TargetDir, "incr", strconv.Itoa(loadedPosition))
		}

		backupPos = loadedPosition + 1
		backupPath = filepath.Join(*TargetDir, "incr/", strconv.Itoa(backupPos))
		err = os.MkdirAll(backupPath, 750)

		if err != nil {
			return errors.New(fmt.Sprintf("[Incremental backup]> Making directories failed, %v", err))
		}

		backupCmd = exec.Command("mariabackup",
			"--host="+*Host,
			"--port="+strconv.Itoa(*Port),
			"--user="+*Username,
			"--password="+*Password,
			"--backup",
			"--version-check",
			"--datadir="+*DataDir,
			"--target_dir="+backupPath,
			"--incremental-basedir="+incrementalBaseDir,
			"--extra-lsndir="+backupPath,
			"--stream=xbstream",
		)

	}

	fmt.Println(backupCmd)
	target = filepath.Join(backupPath, fmt.Sprintf("backup.gz"))
	file, err := os.Create(target)

	if err != nil {
		return err
	}

	defer file.Close()

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	out, err := backupCmd.StdoutPipe()
	backupCmd.Stderr = os.Stderr
	if err != nil {
		return err
	}

	err = backupCmd.Start()

	if err != nil {
		return errors.New(fmt.Sprintf("[Backup]> Failed executing command: %v", err))
	}

	_, err = io.Copy(gzw, out)

	if err != nil {
		return err
	}

	return saveBackupPos(backupPos)
}

func saveBackupPos(position int) error {
	f, err := os.Create(filepath.Join(*TargetDir, "mariabackup.pos"))

	if err != nil {
		return err
	}

	_, err = f.WriteString(strconv.Itoa(position))

	if err != nil {
		return err
	}

	err = f.Close()

	if err != nil {
		return err
	}

	return nil
}
