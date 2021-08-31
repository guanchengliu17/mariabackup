package Manager

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"log"
	"sync/atomic"
)

type UploadProgress struct {
	reader io.Reader
	bytes  int64
}

func (u *UploadProgress) BytesSent() int64 {
	return atomic.LoadInt64(&u.bytes)
}

func (u *UploadProgress) Read(p []byte) (n int, err error) {
	num, err := u.reader.Read(p)
	//Track the number of bytes uploaded
	atomic.AddInt64(&u.bytes, int64(num))

	return num, err
}

func (u *UploadProgress) Upload(sess *session.Session, key string, bucket string, input io.Reader, size int64) (chan ProgressUpdate, error) {
	//Reset the value just in case
	atomic.StoreInt64(&u.bytes, 0)
	u.reader = input
	ul := s3manager.NewUploader(sess)

	//set concurrency
	ul.Concurrency = AwsConcurrencyLevel

	updates := make(chan ProgressUpdate, 32)

	log.Printf("Uploading " + key + " to S3")
	_, err := ul.Upload(&s3manager.UploadInput{
		Body:   u,
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		select {
		case updates <- ProgressUpdate{
			Bytes:    size,
			Total:    u.BytesSent(),
			Finished: false,
			Error:    nil,
		}:
		default:
			log.Printf("Failed to upload" + key + " to S3...")
		}
	}
	log.Printf("Upload finished")
	return updates, nil
}
