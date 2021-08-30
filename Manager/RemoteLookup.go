package Manager

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
)

func RemoteLookup(sess *session.Session, prefix string, bucket string) ([]string, error) {

	client := s3.New(sess)
	out, err := client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int64(100),
		Prefix:  aws.String(prefix),
	})

	if err != nil {
		log.Println("[ERROR] Error during RemoteLookup() error:", err)
		return nil, err
	}
	log.Printf("in remotelookup function")

	results := make([]string, 0)
	for _, record := range out.Contents {
		results = append(results,*record.Key)
	}

	return results, nil
}
