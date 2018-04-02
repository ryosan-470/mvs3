package main

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	// TargetFileName is a file what you want to get from S3
	TargetFileName string
	// OriginBucket is a bucket name saved TargetFileName
	OriginBucket string
	// OriginRegion is AWS S3 region like "us-west-2"
	OriginRegion string
)

func main() {
	downloadFromS3()
}

func downloadFromS3() error {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(OriginRegion),
	}))
	input := &s3.GetObjectInput{
		Bucket: aws.String(OriginBucket),
		Key:    aws.String(TargetFileName),
	}

	f, err := os.Create(TargetFileName)
	if err != nil {
		return err
	}

	downloader := s3manager.NewDownloader(sess)
	n, err := downloader.Download(f, input)
	log.Printf("Successfully download file: %s from S3 (Size: %d B)", TargetFileName, n)
	return err
}
