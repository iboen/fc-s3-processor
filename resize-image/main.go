package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/disintegration/imaging"
)

const (
	AwsRegion = "us-east-1"
)

var awsSession *session.Session

func getAwsSession() *session.Session {
	if awsSession == nil {
		awsSession = session.Must(session.NewSession(&aws.Config{Region: aws.String(AwsRegion)}))
	}
	return awsSession
}

// resizeImage resizes an image to the specified width and height
func resizeImage(data []byte, width int) ([]byte, error) {
	img, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	resizedImg := imaging.Resize(img, width, 0, imaging.Lanczos)
	buf := new(bytes.Buffer)
	err = imaging.Encode(buf, resizedImg, imaging.JPEG)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Handler is the main function triggered by S3 events
func Handler(ctx context.Context, event events.S3Event) error {
	s3Client := s3.New(getAwsSession())

	for _, record := range event.Records {
		bucketName := record.S3.Bucket.Name
		objectKey := record.S3.Object.URLDecodedKey

		fmt.Println("Processing S3 object:", bucketName, objectKey)

		// Get the original image from S3
		getObject, err := s3Client.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objectKey),
		})
		if err != nil {
			return err
		}
		defer getObject.Body.Close()

		imageBytes, err := io.ReadAll(getObject.Body)
		if err != nil {
			return err
		}

		resizeWidth, err := strconv.Atoi(os.Getenv("RESIZE_WIDTH"))
		if err != nil {
			return err
		}

		resizedBytes, err := resizeImage(imageBytes, resizeWidth)
		if err != nil {
			return err
		}

		// Construct a new object key for the resized image
		resizedObjectKey := fmt.Sprintf("resized-%s", objectKey)

		// Upload the resized image to S3 with a new key
		putObject, err := s3Client.PutObject(&s3.PutObjectInput{
			Body:        bytes.NewReader(resizedBytes),
			Bucket:      aws.String(bucketName),
			Key:         aws.String(resizedObjectKey),
			ContentType: aws.String("image/jpeg"),
		})
		if err != nil {
			return err
		}

		fmt.Println("Successfully resized and uploaded:", putObject)
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
