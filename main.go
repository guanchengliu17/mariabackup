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
)

var BackupCommand = flag.NewFlagSet("backup", flag.ExitOnError)
var Host = BackupCommand.String("host", "localhost", "database host")
var Port = BackupCommand.Int("port", 3306, "database port")
var Username = BackupCommand.String("username", "mariabackup", "database username")
var Password = BackupCommand.String("password", "", "database password")
var TargetDir = flag.String("target_dir", "/backup/mariabackup", "directory in which the backups will be placed")
var DataDir = flag.String("datadir", "/var/lib/mysql", "directory where the MySQL data is stored")

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

	createDir(*TargetDir)

	cmd := exec.Command("mariabackup",
		"--host="+*Host,
		"--port="+strconv.Itoa(*Port),
	        "--user="+*Username,
		"--password="+*Password,
	        "--backup",
		"--version-check",
		"--datadir="+*DataDir,
		"--target_dir="+*TargetDir,
		"--extra-lsndir="+*TargetDir,
		"--stream=xbstream",
	)

	target = filepath.Join(target, fmt.Sprintf("backup.gz"))
	file, err := os.Create(target)
	logError(err)
	defer file.Close()

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	out, err := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr
	logError(err)
	cmd.Start()

	io.Copy(gzw, out)
	return nil
}

func createDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0750)
		logError(err)
	}
}

func logError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
