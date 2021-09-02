package Manager

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"log"
	"os"
	"path/filepath"
	"time"
)

type S3Manager struct {
	region     string
	awsSession *session.Session
	accessKey  string
	secret     string
	bucket     string
}

type ProgressUpdate struct {
	Bytes    int64
	Total    int64
	Finished bool
	Error    error
}

func CreateS3Manager(
	AccessKey string,
	Region string,
	Bucket string,
	Secret string,
) (*S3Manager, error) {

	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(AccessKey, Secret, ""),
		Region:      aws.String(Region),
	})

	if err != nil {
		return nil, errors.New("session creation failed")
	}
	return &S3Manager{
		region:     Region,
		accessKey:  AccessKey,
		bucket:     Bucket,
		secret:     Secret,
		awsSession: sess,
	}, nil
}

func (s *S3Manager) Upload(backup string) {

	files := []string{"backup.gz.enc", "xtrabackup_info", "xtrabackup_checkpoints", "checksum"}

	for i := range files {
		ulp := &UploadProgress{}
		fh, err := os.Open(filepath.Join(backup, files[i]))
		if err != nil {
			fmt.Println(err)
		}
		stat, err := fh.Stat()
		if err != nil {
			fmt.Println(err)
		}
		_, err = ulp.Upload(s.awsSession, GenerateUploadS3Path(files[i]), s.bucket, fh, stat.Size())
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (s *S3Manager) Download(backup string, restoreDate string) {
	//check if backup exists in S3
	if s.IsPushed(GenerateDownloadS3Path("backup.gz.enc", restoreDate)) == false {
	}

	files := []string{"backup.gz.enc", "xtrabackup_info", "xtrabackup_checkpoints", "checksum"}

	if _, err := os.Stat(backup); os.IsNotExist(err) {
		err := os.Mkdir(backup, 0755)
		if err != nil {
			log.Printf("Unable to create backups directory")
		}
	}

	for i := range files {
		fh, err := os.OpenFile(filepath.Join(backup, files[i]), os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			fmt.Println(err)
		}

		dlp := &DownloadProgress{}
		_, err = dlp.Download(s.awsSession, GenerateDownloadS3Path(files[i], restoreDate), s.bucket, fh)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (s *S3Manager) IsPushed(backup string) bool {

	results, err := RemoteLookup(s.awsSession, backup, s.bucket)

	if err != nil {
		return false
	}

	if len(results) > 0 {
		return true
	}

	return false
}

func GenerateUploadS3Path(file string) (s3Path string) {

	hostname, _ := os.Hostname()
	currentTime := time.Now()
	date := currentTime.Format("2006-01-02")

	s3Path = filepath.Join(hostname, date, file)

	return s3Path
}

func GenerateDownloadS3Path(file string, restoreDate string) (s3Path string) {

	hostname, _ := os.Hostname()
	s3Path = filepath.Join(hostname, restoreDate, file)

	return s3Path
}
