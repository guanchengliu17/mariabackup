package Manager

import (
        "compress/gzip"
        "errors"
        "fmt"
        "io"
        "os"
        "os/exec"
        "path/filepath"
	"strconv"
	"io/ioutil"
	"log"
)

type RestoreManager struct {
        sourceDirectory string
        targetDirectory string
}

func CreateRestoreManager(
        SourceDirectory string,
        TargetDirectory string) (*RestoreManager, error) {

        return &RestoreManager{
                sourceDirectory: SourceDirectory,
                targetDirectory: TargetDirectory,
        }, nil

}

func (b *RestoreManager) Restore() error {
	log.SetFlags(log.Ldate | log.Ltime)

	err := os.RemoveAll(filepath.Join(b.sourceDirectory, "restore"))
	if err != nil {
		return errors.New(fmt.Sprintf("[Restore backup]> Failed to remove previous backup restore directory, %v", err))
        }

	backupSubDirectory := ""

	backupPosition, _ := b.getBackupPosition()

	for i := 0;  i <= backupPosition; i++ {
		if i == 0 {
			backupSubDirectory = "full"
		} else {
			backupSubDirectory = filepath.Join("incr", strconv.Itoa(i))
		}

		log.Println("Decompressing", filepath.Join(filepath.Join(b.sourceDirectory, backupSubDirectory), "backup.gz"), "to", filepath.Join(b.sourceDirectory, "restore", backupSubDirectory))
		err := b.decompressBackup(backupSubDirectory)

		if err != nil {
			return err
		}
	}
	return nil
}

func (b *RestoreManager) decompressBackup(backupSubDirectory string) error {
	backupRestoreTemporaryDirectory := filepath.Join(b.sourceDirectory, "restore", backupSubDirectory)

	err := os.MkdirAll(backupRestoreTemporaryDirectory, 750)

	if err != nil {
		return errors.New(fmt.Sprintf("[RestoreManager]> Making directories failed, %v", err))
	}

	f, err := os.Open(filepath.Join(filepath.Join(b.sourceDirectory, backupSubDirectory), "backup.gz"))

        if err != nil {
                return err
        }

        defer f.Close()

        gzr, err := gzip.NewReader(f)

        if err != nil {
                return err
        }

        defer gzr.Close()

        command := exec.Command("mbstream", "-x", "-C", backupRestoreTemporaryDirectory)

        out, err := command.StdinPipe()
        command.Stderr = os.Stderr
        if err != nil {
                return err
        }

        err = command.Start()

        if err != nil {
                return errors.New(fmt.Sprintf("[RestoreManager Restore()]> Failed executing command: %v", err))
        }

        _, err = io.Copy(out, gzr)

        if err != nil {
                return err
        }

	return nil
}

func (b *RestoreManager) getBackupPosition() (int, error) {
	data, err := ioutil.ReadFile(filepath.Join(b.sourceDirectory, "mariabackup.pos"))
        if err != nil {
		return 0, errors.New(fmt.Sprintf("[RestoreManager]> Failed to read backup position file, %v", err))
	}

	backupPosition, err := strconv.Atoi(string(data))

	if err != nil {
                return 0, err
        }

	return backupPosition, nil
}
