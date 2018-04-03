package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/mholt/archiver"
	"github.com/yeka/zip" // Fork of Go's archive/zip to add reading/writing of password protected zip files.
)

var (
	// TargetFileName is a file what you want to get from S3
	TargetFileName string
	// OriginBucket is a bucket name saved TargetFileName
	OriginBucket string
	// OriginRegion is AWS S3 region like "us-west-2"
	OriginRegion string
	// TargetFileName password if it is encrypted
	Password string
)

func main() {
	targzList, _ := unzipWithPassword()
	fmt.Printf("%v\n", targzList)
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

func unzipWithPassword() ([]string, error) {
	r, err := zip.OpenReader(TargetFileName)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	compressedFileList := make([]string, 0)

	for _, f := range r.File {
		if f.IsEncrypted() {
			f.SetPassword(Password)
		}

		r, err := f.Open()
		if err != nil {
			return nil, err
		}
		buf, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		ioutil.WriteFile(f.Name, buf, 0644)
		fmt.Printf("extract file %s %d B\n", f.Name, len(buf))
		compressedFileList = append(compressedFileList, f.Name)
		defer r.Close()
	}
	return compressedFileList, nil
}

func extractTarGz(targzList []string) {
	for _, targz := range targzList {
		archiver.TarGz.Open(targz, "tmp")
	}
}
