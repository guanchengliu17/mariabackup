package main

import (
    "io"
    "os"
    "os/exec"
    "compress/gzip"
    "fmt"
    "path/filepath"
    "flag"
    "strconv"
    "log"
    "io/ioutil"
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

	switch os.Args[1] {
	case "backup":
		BackupCommand.Parse(os.Args[2:])
	default:
		fmt.Printf("%q is not valid command.\n", os.Args[1])
		os.Exit(2)
	}

	if BackupCommand.Parsed() {
		fmt.Println("Backup data directory:", *DataDir)
		fmt.Println("Backup target directory:", *TargetDir)
		backup(*TargetDir)
	}

}

func backup(target string) (error) {

	var backupPath string
	var backupCmd *exec.Cmd
	var backupPos string
	var incrementalBaseDir string

	if *Type == "full" {
		os.RemoveAll(*TargetDir)
		backupPath = filepath.Join(*TargetDir, "full")
		os.MkdirAll(backupPath, 750)
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
		backupPos = "0"
	} else if *Type == "incr" {

		file, err := ioutil.ReadFile(filepath.Join(*TargetDir, "mariabackup.pos"))
		logError(err)
		pos, err := strconv.Atoi(string(file))
		if pos == 0 {
			incrementalBaseDir = filepath.Join(*TargetDir, "full")
		} else {
			incrementalBaseDir = filepath.Join(*TargetDir, "incr", strconv.Itoa(pos))
		}
		pos = pos + 1
		backupPos = strconv.Itoa(pos)
		backupPath = filepath.Join(*TargetDir, "incr/", backupPos)
		os.MkdirAll(backupPath, 750)
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
	logError(err)
	defer file.Close()

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	out, err := backupCmd.StdoutPipe()
	backupCmd.Stderr = os.Stderr
	logError(err)
	backupCmd.Start()

	io.Copy(gzw, out)
	saveBackupPos(backupPos)
	return nil
}

func saveBackupPos(state string) {
    f, err := os.Create(filepath.Join(*TargetDir, "mariabackup.pos"))
    logError(err)
    l, err := f.WriteString(state)
    logError(err)
    fmt.Println(l, "bytes written successfully")
    err = f.Close()
    logError(err)
}

func logError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
