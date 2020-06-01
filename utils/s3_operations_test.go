package utils

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"testing"
)

func TestS3_DeleteFolder(t *testing.T) {

	svc := s3.New(session.New())
	bucket := "ph-stream"
	path := "test/jeorch-test"
	
	type args struct {
		svc    *s3.S3
		bucket string
		path   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{name:"t001", args:args{
			svc:    svc,
			bucket: bucket,
			path:   path,
		}, wantErr:false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := S3_DeleteFolder(tt.args.svc, tt.args.bucket, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("S3_DeleteFolder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}