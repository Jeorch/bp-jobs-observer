package utils

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	S3_PathPrefix = "s3a://"
)

func S3_ListObjects(svc *s3.S3, bucket string, subPath string) ([]*s3.Object, error) {

	input := &s3.ListObjectsInput{
		Bucket:  aws.String(bucket),
		Prefix: aws.String(subPath),
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
		return nil, err
	}

	return result.Contents, err

}

func S3_IsEmptyObject(obj *s3.Object) bool {
	return *obj.Size == 0
}

func S3_DeleteObject(svc *s3.S3, bucket string, key string) error {

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := svc.DeleteObject(input)
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
		return err
	}
	return nil
}

func S3_DeleteFolder(svc *s3.S3, bucket string, path string) error {

	objs, err := S3_ListObjects(svc, bucket, path)
	if err != nil {
		return err
	}
	for _, obj := range objs {
		err = S3_DeleteObject(svc, bucket, *obj.Key)
		if err != nil {
			return err
		}
	}
	return nil
}
