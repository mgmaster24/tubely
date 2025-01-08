package main

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func generatePresignedURL(
	s3Client *s3.Client,
	bucket, key string,
	expireTime time.Duration,
) (string, error) {
	presignClient := s3.NewPresignClient(s3Client)
	getObjInput := s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}
	presignedReq, err := presignClient.PresignGetObject(
		context.Background(),
		&getObjInput,
		s3.WithPresignExpires(expireTime),
	)
	if err != nil {
		return "", err
	}

	return presignedReq.URL, nil
}
