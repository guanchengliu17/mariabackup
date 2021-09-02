package Manager

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"log"
	"sync/atomic"
)

type DownloadProgress struct {
	writer io.WriterAt
	bytes  int64
}

func (d *DownloadProgress) Download(sess *session.Session, key string, bucket string, writer io.WriterAt) (chan ProgressUpdate, error) {
	//Resets the value just in case
	atomic.StoreInt64(&d.bytes, 0)

	updates := make(chan ProgressUpdate, 32)
	d.writer = writer

	//Create s3 client and determine file size
	s3Client := s3.New(sess)
	head, err := s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, err
	}

	//determine total size of the file
	fileSize := head.ContentLength

	//create download manager
	dl := s3manager.NewDownloader(sess)
	dl.Concurrency = AwsConcurrencyLevel
	log.Printf("Downloading " + key + " from S3")
	_, err = dl.Download(d, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		select {
		case updates <- ProgressUpdate{
			Bytes:    *fileSize,
			Total:    d.BytesWritten(),
			Finished: false,
			Error:    err,
		}:
		default:
			log.Printf("Failed to download" + key + " from S3...")
		}
	}

	return updates, nil
}

func (d *DownloadProgress) WriteAt(p []byte, off int64) (n int, err error) {
	atomic.AddInt64(&d.bytes, int64(len(p))) //increment

	return d.writer.WriteAt(p, off)
}

func (d *DownloadProgress) BytesWritten() int64 {
	return atomic.LoadInt64(&d.bytes)
}
