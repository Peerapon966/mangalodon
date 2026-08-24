package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Peerapon966/blackbox/scraper/internal/apperr"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

// ================== Public Types ==================

type Client struct {
	*s3.Client
}

type Mode string

const (
	None        Mode = ""
	IfMatch     Mode = "IfMatch"
	IfNoneMatch Mode = "IfNoneMatch"
)

type UploadObjectInput struct {
	ObjectKey string
	Data      []byte
	Mode      Mode
	ETag      *string
}

type DownloadObjectOutput struct {
	Data []byte
	ETag *string
}

// ================== Constructor ==================

func NewClient(ctx context.Context) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return nil, &apperr.ScraperError{
			Code:    apperr.EnvVarNotSet,
			Message: "unable to load SDK config.",
			Err:     err.Error(),
		}
	}

	s3Client := s3.NewFromConfig(cfg)

	return &Client{
		s3Client,
	}, nil
}

// ================== Public Methods ==================

func (c *Client) UploadObject(ctx context.Context, params *UploadObjectInput) error {
	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		return &apperr.ScraperError{
			Code:    apperr.EnvVarNotSet,
			Message: "BUCKET_NAME environment variable must be set.",
		}
	}

	reader := bytes.NewReader(params.Data)
	var result *s3.PutObjectOutput
	var err error
	var errCode apperr.ErrorCode
	var errMsg string
	switch params.Mode {
	case IfMatch:
		errCode = apperr.IfMatchPreconditionFailed
		errMsg = fmt.Sprintf("Error while uploading object %s to %s. Mid-air collision detected. File changed on server.", params.ObjectKey, bucketName)
		result, err = c.PutObject(ctx, &s3.PutObjectInput{
			Bucket:  aws.String(bucketName),
			Key:     aws.String(params.ObjectKey),
			Body:    reader,
			IfMatch: params.ETag,
		})
	case IfNoneMatch:
		errCode = apperr.IfNoneMatchPreconditionFailed
		errMsg = fmt.Sprintf("Error while uploading object %s to %s. File already exists. Upload aborted.", params.ObjectKey, bucketName)
		result, err = c.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(bucketName),
			Key:         aws.String(params.ObjectKey),
			Body:        reader,
			IfNoneMatch: aws.String("*"),
		})
	case None:
		result, err = c.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(params.ObjectKey),
			Body:   reader,
		})
	default:
		return &apperr.ScraperError{
			Code:    apperr.InvalidPutObjectMode,
			Message: fmt.Sprintf("Invalid mode. PutObject mode must be 'IfMatch', 'IfNoneMatch', or 'None', %s given.", params.Mode),
		}
	}
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "PreconditionFailed" {
			return &apperr.ScraperError{
				Code:    errCode,
				Message: errMsg,
				Err:     err.Error(),
			}
		} else {
			return &apperr.ScraperError{
				Code:    apperr.PutObjectFailed,
				Message: fmt.Sprintf("Couldn't upload object %s to %s.", params.ObjectKey, bucketName),
				Err:     err.Error(),
			}
		}
	} else {
		err = s3.NewObjectExistsWaiter(c).Wait(
			ctx, &s3.HeadObjectInput{Bucket: aws.String(bucketName), Key: aws.String(params.ObjectKey), IfMatch: result.ETag}, time.Minute)
		if err != nil {
			return &apperr.ScraperError{
				Code:    apperr.PutObjectFailed,
				Message: fmt.Sprintf("Failed attempt to wait for object %s to exist.\n", params.ObjectKey),
				Err:     err.Error(),
			}
		}
	}

	return err
}

func (c *Client) DownloadObject(ctx context.Context, objectKey string) (DownloadObjectOutput, error) {
	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		return DownloadObjectOutput{}, &apperr.ScraperError{
			Code:    apperr.EnvVarNotSet,
			Message: "BUCKET_NAME environment variable must be set.",
		}
	}

	result, err := c.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			return DownloadObjectOutput{}, &apperr.ScraperError{
				Code:    apperr.NoSuchKey,
				Message: fmt.Sprintf("Couldn't get object %s from bucket %s. No such key exists.\n", objectKey, bucketName),
				Err:     err.Error(),
			}
		} else {
			return DownloadObjectOutput{}, &apperr.ScraperError{
				Code:    apperr.GetObjectFailed,
				Message: fmt.Sprintf("Couldn't get object %v from %v.", bucketName, objectKey),
				Err:     err.Error(),
			}
		}
	}

	defer result.Body.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		return DownloadObjectOutput{}, &apperr.ScraperError{
			Code:    apperr.GetObjectFailed,
			Message: fmt.Sprintf("Couldn't read object body from %v.", objectKey),
			Err:     err.Error(),
		}
	}

	return DownloadObjectOutput{
		Data: body,
		ETag: result.ETag,
	}, nil
}

func (c *Client) RemoveObject(ctx context.Context, objectKey string) error {
	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		return &apperr.ScraperError{
			Code:    apperr.EnvVarNotSet,
			Message: "BUCKET_NAME environment variable must be set.",
		}
	}

	_, err := c.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return &apperr.ScraperError{
			Code:    apperr.DeleteObjectFailed,
			Message: fmt.Sprintf("Couldn't delete object %v from %v.", bucketName, objectKey),
			Err:     err.Error(),
		}
	}

	return nil
}

func (c *Client) CheckObjectExists(ctx context.Context, objectKey string) (bool, error) {
	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		return false, &apperr.ScraperError{
			Code:    apperr.EnvVarNotSet,
			Message: "BUCKET_NAME environment variable must be set.",
		}
	}

	_, err := c.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		var httpErr *http.ResponseError
		if errors.As(err, &httpErr) && httpErr.Response.StatusCode == 404 {
			return false, nil
		}

		return false, &apperr.ScraperError{
			Code:    apperr.HeadObjectFailed,
			Message: fmt.Sprintf("Couldn't head object %v from %v.", bucketName, objectKey),
			Err:     err.Error(),
		}
	}

	return true, nil
}
