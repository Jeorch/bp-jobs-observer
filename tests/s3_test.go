package tests

import (
	"fmt"
	"github.com/PharbersDeveloper/bp-jobs-observer/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"testing"
)
import "github.com/aws/aws-sdk-go/service/s3"

func TestReadS3(t *testing.T) {

	svc := s3.New(session.New())
	input := &s3.ListObjectsInput{
		Bucket:  aws.String("ph-stream"),
		Prefix: aws.String("test/"),
	}

	result, err := svc.ListObjects(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				fmt.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	for _, obj := range result.Contents {
		fmt.Println(obj)
		fmt.Println(utils.S3_IsEmptyObject(obj))
	}

}

func TestDeleteObject(t *testing.T) {
	svc := s3.New(session.New())
	input := &s3.DeleteObjectInput{
		Bucket: aws.String("ph-stream"),
		Key:    aws.String("test/000/"),
	}

	result, err := svc.DeleteObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(result)
}
