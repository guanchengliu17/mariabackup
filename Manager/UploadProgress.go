package Manager

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"sync/atomic"
	"time"
)

/**
S3 upload implementation with progression tracking

Usage:
	Call Upload()
	Read the ProgressUpdate channel until the channel is closed with error or ProgressUpdate.Finished = true
*/
type UploadProgress struct {
	reader io.Reader
	bytes  int64
}

const AwsConcurrencyLevel = 16

/**
Number of bytes uploaded
*/
func (u *UploadProgress) BytesSent() int64 {
	return atomic.LoadInt64(&u.bytes)
}

/**
Wrapper for Read() for compatibility with io.Reader
*/
func (u *UploadProgress) Read(p []byte) (n int, err error) {
	num, err := u.reader.Read(p)
	//Track the number of bytes uploaded
	atomic.AddInt64(&u.bytes, int64(num))

	return num, err
}

/**
NOTE: This function should not be called concurrently, create separate instance instead
Uploads file to S3, input file handle and also size of the file in byte to report total bytes in the ProgressUpdate channel
*/
func (u *UploadProgress) Upload(sess *session.Session, key string, bucket string, input io.Reader, size int64) (chan ProgressUpdate, error) {
	//Reset the value just in case
	atomic.StoreInt64(&u.bytes, 0)

	u.reader = input
	ul := s3manager.NewUploader(sess)

	//set concurrency
	ul.Concurrency = AwsConcurrencyLevel

	updates := make(chan ProgressUpdate, 32)
	exit := make(chan bool, 1)

	go func() {
		defer func() {
			exit <- true
		}()

		_, err := ul.Upload(&s3manager.UploadInput{
			Body:   u,
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})

		if err != nil {
			select {
			case updates <- ProgressUpdate{
				Bytes:    0,
				Total:    0,
				Finished: false,
				Error:    nil,
			}:
			default:
			}
		}

	}()

	go func() {
		finished := false
		for {

			select {
			case <-exit:
				finished = true
			default:
			}
			time.Sleep(time.Second) //update rate 1hz

			updates <- ProgressUpdate{
				Bytes:    size,
				Total:    u.BytesSent(),
				Finished: finished,
				Error:    nil,
			}

			if finished {
				close(updates)
				return
			}

		}

	}()

	return updates, nil
}
