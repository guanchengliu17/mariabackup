package Manager

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"os"
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

func CreateUploadManager(
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
		return nil, errors.New("upload manager initialization failed")
	}
	return &S3Manager{
		region:     Region,
		accessKey:  AccessKey,
		bucket:     Bucket,
		secret:     Secret,
		awsSession: sess,
	}, nil
}

func (s *S3Manager) Upload(backup string) (chan ProgressUpdate, error) {

	ulp := &UploadProgress{}

	fh, err := os.Open(backup)
	if err != nil {
		return nil, err
	}

	stat, err := fh.Stat()
	if err != nil {
		return nil, err
	}

	ul, err := ulp.Upload(s.awsSession, backup, s.bucket, fh, stat.Size())

	if err != nil {
		return nil, err
	}

	return ul, nil
}
