package Manager

import (
	"errors"
	"fmt"
	gzip "github.com/klauspost/pgzip"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
)

type RestoreManager struct {
	sourceDirectory    string
	targetDirectory    string
	workDirectory      string
	mariaBackupBinary  string
	backupPositionFile string
	mbStreamBinary     string
	gzBlockSize        int
	gzThreads          int
}

func CreateRestoreManager(SourceDirectory string, TargetDirectory string, WorkDirectory string, MariaBackupBinary string, BackupPositionFile string, MbStreamBinary string, CompressionBlockSize int, DecompressionThreads int, ) (*RestoreManager, error) {

	return &RestoreManager{
		sourceDirectory:    SourceDirectory,
		targetDirectory:    TargetDirectory,
		workDirectory:      WorkDirectory,
		mariaBackupBinary:  MariaBackupBinary,
		backupPositionFile: BackupPositionFile,
		mbStreamBinary:     MbStreamBinary,
		gzBlockSize:        CompressionBlockSize,
		gzThreads:          DecompressionThreads,
	}, nil

}

func (b *RestoreManager) Restore() error {
	f, err := os.Open(b.targetDirectory)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Readdir(1)

	if err != io.EOF {
		return errors.New(fmt.Sprintf("[Restore backup]> Target directory %v is not empty", b.targetDirectory))
	}

	err = os.RemoveAll(b.workDirectory)
	if err != nil {
		return errors.New(fmt.Sprintf("[Restore backup]> Failed to remove previous backup restore directory, %v", err))
	}

	backupSubDirectory := ""

	backupPosition, _ := b.getBackupPosition()

	for i := 0; i <= backupPosition; i++ {
		if i == 0 {
			backupSubDirectory = "full"
		} else {
			backupSubDirectory = filepath.Join("incr", strconv.Itoa(i))
		}

		log.Println("Decompressing", filepath.Join(filepath.Join(b.sourceDirectory, backupSubDirectory), "backup.gz"), "to", filepath.Join(b.workDirectory, backupSubDirectory))
		err := b.decompressBackup(backupSubDirectory)

		if err != nil {
			return err
		}

		log.Println("Preparing", filepath.Join(b.workDirectory, backupSubDirectory))
		err = b.prepareBackup(backupSubDirectory)

		if err != nil {
			return err
		}
	}

	err = b.moveBackupToTargetDirectory()

	if err != nil {
		return err
	}

	return nil
}

func (b *RestoreManager) decompressBackup(backupSubDirectory string) error {
	workDirectory := filepath.Join(b.workDirectory, backupSubDirectory)

	err := os.MkdirAll(workDirectory, 750)

	if err != nil {
		return errors.New(fmt.Sprintf("[RestoreManager]> Making directories failed, %v", err))
	}

	f, err := os.Open(filepath.Join(filepath.Join(b.sourceDirectory, backupSubDirectory), "backup.gz"))

	if err != nil {
		return err
	}

	defer f.Close()

	gzr, err := gzip.NewReaderN(f, b.gzBlockSize, b.gzThreads)

	if err != nil {
		return err
	}

	defer gzr.Close()

	command := exec.Command(b.mbStreamBinary, "-x", "-C", workDirectory)

	out, err := command.StdinPipe()
	command.Stderr = os.Stderr
	if err != nil {
		return err
	}

	err = command.Start()

	if err != nil {
		return errors.New(fmt.Sprintf("[RestoreManager Restore()]> Failed executing mbstream command: %v", err))
	}

	_, err = io.Copy(out, gzr)

	if err != nil {
		return err
	}

	return nil
}

func (b *RestoreManager) prepareBackup(backupSubDirectory string) error {
	command := exec.Command(b.mariaBackupBinary,
		"--prepare",
		"--target-dir="+filepath.Join(b.workDirectory, "full"),
	)

	if backupSubDirectory != "full" {
		command.Args = append(command.Args, "--incremental-basedir="+filepath.Join(b.workDirectory, backupSubDirectory))
	}

	command.Stderr = os.Stderr
	command.Stdout = os.Stdout
	err := command.Start()

	if err != nil {
		return errors.New(fmt.Sprintf("[RestoreManager Restore()]> Failed executing mariabackup --prepare command: %v", err))
	}

	err = command.Wait()

	if err != nil {
		return err
	}

	//check if the exit code was 0
	exitCode := command.ProcessState.ExitCode()

	if exitCode != 0 {
		return errors.New("Failed to prepare backup, exit code:" + strconv.Itoa(exitCode))
	}

	return nil
}

func (b *RestoreManager) moveBackupToTargetDirectory() error {
	command := exec.Command(b.mariaBackupBinary,
		"--move-back",
		"--target-dir="+filepath.Join(b.workDirectory, "full"),
	)

	command.Stderr = os.Stderr
	command.Stdout = os.Stdout
	err := command.Start()

	if err != nil {
		return errors.New(fmt.Sprintf("[RestoreManager Restore()]> Failed executing mariabackup --move-back command: %v", err))
	}

	err = command.Wait()

	if err != nil {
		return err
	}

	//check if the exit code was 0
	exitCode := command.ProcessState.ExitCode()

	if exitCode != 0 {
		return errors.New("Failed to move backup to target directory, exit code:" + strconv.Itoa(exitCode))
	}

	group, err := user.Lookup("mysql")

	if err != nil {
		return err
	}

	uid, _ := strconv.Atoi(group.Uid)
	gid, _ := strconv.Atoi(group.Gid)

	err = filepath.Walk(b.targetDirectory, func(name string, f os.FileInfo, err error) error {
		if err == nil {
			err = os.Chown(name, uid, gid)
		}
		return err
	})

	return nil
}

func (b *RestoreManager) getBackupPosition() (int, error) {
	data, err := ioutil.ReadFile(b.backupPositionFile)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("[RestoreManager]> Failed to read backup position file, %v", err))
	}

	backupPosition, err := strconv.Atoi(string(data))

	if err != nil {
		return 0, err
	}

	return backupPosition, nil
}
