package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/mholt/archiver"
	"github.com/yeka/zip" // Fork of Go's archive/zip to add reading/writing of password protected zip files.
)

var (
	// OriginFileName is a file what you want to get from S3
	OriginFileName string = os.Getenv("ORIGIN_FILENAME")
	// OriginBucket is a bucket name saved TargetFileName
	OriginBucket string = os.Getenv("ORIGIN_BUCKET")
	// OriginRegion is AWS S3 region like "us-west-2"
	OriginRegion string = os.Getenv("ORIGIN_REGION")
	// TargetFileName password if it is encrypted
	Password string = os.Getenv("PASSWORD") // TODO: Use KMS
	// TargetBucket is a bucket name to upload document which downloaded from OriginBucket
	TargetBucket string = os.Getenv("TARGET_BUCKET")
	// TargetRegion is AWS S3 region like "ap-northeast-1"
	TargetRegion string = os.Getenv("TARGET_REGION")
)

func main() {
	lambda.Start(run)
}

func run() {
	targzList, _ := unzipWithPassword()
	log.Printf("Unzipped: %v\n", targzList)
	savedDir, err := ioutil.TempDir("", "doc")
	if err != nil {
		log.Fatalln("Failed to create temp dir")
	}
	extractTarGz(targzList, savedDir)
	err = uploadToS3(savedDir)
	if err != nil {
		log.Fatalf("Failed to upload %v to S3", targzList)
	}
}

func downloadFromS3() error {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(OriginRegion),
	}))
	input := &s3.GetObjectInput{
		Bucket: aws.String(OriginBucket),
		Key:    aws.String(OriginFileName),
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

func extractTarGz(targzList []string, savedDir string) {
	for _, targz := range targzList {
		archiver.TarGz.Open(targz, savedDir)
		log.Printf("Extract %s to %s\n", targz, savedDir)
	}
}

func uploadToS3(dir string) error {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(TargetRegion),
	}))
	uploader := s3manager.NewUploader(sess)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", dir, err)
			return err
		}
		filename := path[len(dir):]
		f, err := os.Open(path)
		stat, err := f.Stat()
		if !stat.IsDir() {
			result, err := uploader.Upload(&s3manager.UploadInput{
				Bucket: aws.String(TargetBucket),
				Key:    aws.String(filename),
				Body:   f,
			})
			if err != nil {
				return fmt.Errorf("Failed to upload file %v", err)
			}
			log.Printf("Upload: %s\n", result.Location)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", dir, err)
	}
	return nil
}
