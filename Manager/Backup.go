package Manager

import (
	"errors"
	"fmt"
	gzip "github.com/klauspost/pgzip"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

const (
	FullBackupMode        = "full"
	IncrementalBackupMode = "incremental"
)

type BackupManager struct {
	targetDirectory    string
	host               string
	port               int
	username           string
	password           string
	mode               string
	dataDirectory      string
	mariabackupBinary  string
	backupPositionFile string
}

func CreateBackupManager(
	TargetDirectory string,
	Host string,
	Port int,
	Username string,
	Password string,
	Mode string,
	DataDirectory string,
	MariabackupBinary string,
	BackupPositionFile string) (*BackupManager, error) {

	switch Mode {
	case FullBackupMode, IncrementalBackupMode:
		break
	default:
		return nil, errors.New("invalid mode only ´full´ or ´incremental´ are supported, got: " + Mode)
	}

	return &BackupManager{
		targetDirectory:    TargetDirectory,
		host:               Host,
		port:               Port,
		username:           Username,
		password:           Password,
		mode:               Mode,
		dataDirectory:      DataDirectory,
		mariabackupBinary:  MariabackupBinary,
		backupPositionFile: BackupPositionFile,
	}, nil

}

func (b *BackupManager) Backup() error {

	backupPath := ""
	backupPos := 0
	incrementalBaseDir := ""

	if b.mode == FullBackupMode {
		err := os.RemoveAll(b.targetDirectory)
		if err != nil {
			return errors.New(fmt.Sprintf("[Full backup]> Failed to remove targetDir, %v", err))
		}

		backupPath = filepath.Join(b.targetDirectory, "full")
		err = os.MkdirAll(backupPath, 750)

		if err != nil {
			return errors.New(fmt.Sprintf("[Full backup]> Making directories failed, %v", err))
		}

	} else if b.mode == IncrementalBackupMode {

		data, err := ioutil.ReadFile(b.backupPositionFile)

		if err != nil {
			return errors.New(fmt.Sprintf("[Incremental backup]> Failed to read backup position file, %v", err))
		}

		loadedPosition, err := strconv.Atoi(string(data))

		if loadedPosition == 0 {
			incrementalBaseDir = filepath.Join(b.targetDirectory, "full")
		} else {
			incrementalBaseDir = filepath.Join(b.targetDirectory, "incr", strconv.Itoa(loadedPosition))
		}

		backupPos = loadedPosition + 1
		backupPath = filepath.Join(b.targetDirectory, "incr/", strconv.Itoa(backupPos))
		err = os.MkdirAll(backupPath, 750)

		if err != nil {
			return errors.New(fmt.Sprintf("[Incremental backup]> Making directories failed, %v", err))
		}
	}

	command := exec.Command(b.mariabackupBinary,
		"--host="+b.host,
		"--port="+strconv.Itoa(b.port),
		"--user="+b.username,
		"--password="+b.password,
		"--backup",
		"--version-check",
		"--datadir="+b.dataDirectory,
		"--target_dir="+backupPath,
		"--extra-lsndir="+backupPath,
		"--parallel=4",
		"--stream=xbstream",
	)

	if len(incrementalBaseDir) > 0 {
		command.Args = append(command.Args, "--incremental-basedir="+incrementalBaseDir)
	}

	err := b.executeCommandAndSaveOutput(backupPath, command)

	if err != nil {
		return err
	}

	return b.saveBackupPosition(backupPos, b.targetDirectory)
}

func (b *BackupManager) executeCommandAndSaveOutput(backupPath string, command *exec.Cmd) error {

	file, err := os.Create(filepath.Join(backupPath, "backup.gz"))

	if err != nil {
		return err
	}

	defer file.Close()

	gzw, err := gzip.NewWriterLevel(file, gzip.BestSpeed)

	gzw.SetConcurrency(2048<<10, 16) //todo: implement config, handle errors

	if err != nil {
		return errors.New("Failed to create gzip writer:" + err.Error())
	}

	defer gzw.Close()

	out, err := command.StdoutPipe()
	command.Stderr = os.Stderr
	if err != nil {
		return err
	}

	err = command.Start()

	if err != nil {
		return errors.New(fmt.Sprintf("[BackupManager Backup()]> Failed executing command: %v", err))
	}

	_, err = io.Copy(gzw, out)

	if err != nil {
		return err
	}

	err = command.Wait()

	if err != nil {
		return err
	}

	//check if the exit code was 0
	exitCode := command.ProcessState.ExitCode()

	if exitCode != 0 {
		return errors.New("Backup failed, exit code:" + strconv.Itoa(exitCode))
	}

	return nil
}

func (b *BackupManager) saveBackupPosition(position int, targetDirectory string) error {
	f, err := os.Create(b.backupPositionFile)

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
