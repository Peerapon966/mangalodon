package task

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Peerapon966/blackbox/scraper/internal/apperr"
	"github.com/Peerapon966/blackbox/scraper/internal/crypto"
	"github.com/Peerapon966/blackbox/scraper/internal/s3"
	"github.com/Peerapon966/blackbox/scraper/internal/utils"
)

type dataType string

const (
	activeTaskIndex  dataType = "active task index"
	activeTaskState  dataType = "active task state"
	historyTaskIndex dataType = "history task index"
	library          dataType = "library"
)

type loadDataInput[T any] struct {
	ctx      context.Context
	s3Client *s3.Client
	dType    dataType
	objkey   string
	zero     T
	secret   []byte
}

type pushDataInput[T any] struct {
	ctx      context.Context
	s3Client *s3.Client
	dType    dataType
	objkey   string
	data     T
	etag     *string
	mode     s3.Mode
	secret   []byte
}

func loadData[T any](params loadDataInput[T]) (T, *string, error) {
	obj, err := params.s3Client.DownloadObject(params.ctx, params.objkey)
	if err != nil {
		var scraperErr *apperr.ScraperError
		if errors.As(err, &scraperErr) && scraperErr.Code == apperr.NoSuchKey {
			return params.zero, nil, nil
		} else {
			return params.zero, nil, err
		}
	}

	decData, err := crypto.DecryptBlob(obj.Data, params.secret)
	if err != nil {
		return params.zero, nil, err
	}

	var data T
	err = json.Unmarshal(decData, &data)
	if err != nil {
		return params.zero, nil, &apperr.ScraperError{
			Code:    apperr.UnmarshalFailed,
			Message: fmt.Sprintf("Couldn't unmarshal %s.", strings.ToLower(string(params.dType))),
			Err:     err.Error(),
		}
	}

	return data, obj.ETag, nil
}

func pushData[T any](params pushDataInput[T]) error {
	data, err := json.Marshal(params.data)
	if err != nil {
		return &apperr.ScraperError{
			Code:    apperr.MarshalFailed,
			Message: fmt.Sprintf("Couldn't marshal %s.", strings.ToLower(string(params.dType))),
			Err:     err.Error(),
		}
	}

	encData, err := crypto.EncryptBlob(data, params.secret)
	if err != nil {
		return err
	}

	var mode s3.Mode
	if params.mode == s3.None {
		objExists, err := params.s3Client.CheckObjectExists(params.ctx, params.objkey)
		if err != nil {
			return err
		}

		mode = utils.If(objExists, s3.IfMatch, s3.IfNoneMatch)
	} else {
		mode = params.mode
	}
	err = params.s3Client.UploadObject(params.ctx, &s3.UploadObjectInput{
		ObjectKey: params.objkey,
		Data:      encData,
		Mode:      mode,
		ETag:      params.etag,
	})

	return err
}

func loadActiveTaskIndex(ctx context.Context, s3Client *s3.Client, secrets crypto.Secrets) (ActiveTaskIndex, *string, error) {
	key := "data/tasks/active.json.enc"

	return loadData(loadDataInput[ActiveTaskIndex]{
		ctx:      ctx,
		s3Client: s3Client,
		dType:    activeTaskIndex,
		objkey:   key,
		zero:     ActiveTaskIndex{},
		secret:   secrets.DEK,
	})
}

func pushActiveTaskIndex(ctx context.Context, s3Client *s3.Client, secrets crypto.Secrets, data ActiveTaskIndex, etag *string) error {
	key := "data/tasks/active.json.enc"

	return pushData(pushDataInput[ActiveTaskIndex]{
		ctx:      ctx,
		s3Client: s3Client,
		dType:    activeTaskIndex,
		objkey:   key,
		data:     data,
		etag:     etag,
		secret:   secrets.DEK,
	})
}

func loadActiveTaskState(ctx context.Context, s3Client *s3.Client, secrets crypto.Secrets, id string) (TaskState, *string, error) {
	key := fmt.Sprintf("data/tasks/status/%s.json.enc", id)

	taskstate, etag, err := loadData(loadDataInput[TaskState]{
		ctx:      ctx,
		s3Client: s3Client,
		dType:    activeTaskIndex,
		objkey:   key,
		zero:     TaskState{},
		secret:   secrets.DEK,
	})
	// should return error instead of zero when task state not found
	if err == nil && taskstate.Status == Unknown {
		return TaskState{}, nil, &apperr.ScraperError{
			Code:    apperr.NoSuchKey,
			Message: fmt.Sprintf("Can't get object %s from bucket %s. No such key exists.\n", key, os.Getenv("BUCKET_NAME")),
		}
	}

	return taskstate, etag, err
}

func pushActiveTaskState(ctx context.Context, s3Client *s3.Client, secrets crypto.Secrets, id string, data TaskState, etag *string, mode s3.Mode) error {
	key := fmt.Sprintf("data/tasks/status/%s.json.enc", id)

	return pushData(pushDataInput[TaskState]{
		ctx:      ctx,
		s3Client: s3Client,
		dType:    library,
		objkey:   key,
		data:     data,
		etag:     etag,
		mode:     mode,
		secret:   secrets.DEK,
	})
}

func loadHistoryTaskIndex(ctx context.Context, s3Client *s3.Client, secrets crypto.Secrets, year int, month int) (HistoryTaskIndex, *string, error) {
	key := fmt.Sprintf("data/tasks/history/%04d-%02d.json.enc", year, month)

	return loadData(loadDataInput[HistoryTaskIndex]{
		ctx:      ctx,
		s3Client: s3Client,
		dType:    historyTaskIndex,
		objkey:   key,
		zero:     HistoryTaskIndex{},
		secret:   secrets.DEK,
	})
}

func pushHistoryTaskIndex(ctx context.Context, s3Client *s3.Client, secrets crypto.Secrets, data HistoryTaskIndex, etag *string, year int, month int) error {
	key := fmt.Sprintf("data/tasks/history/%04d-%02d.json.enc", year, month)

	return pushData(pushDataInput[HistoryTaskIndex]{
		ctx:      ctx,
		s3Client: s3Client,
		dType:    historyTaskIndex,
		objkey:   key,
		data:     data,
		etag:     etag,
		secret:   secrets.DEK,
	})
}

func loadLibrary(ctx context.Context, s3Client *s3.Client, secrets crypto.Secrets) (Library, *string, error) {
	zero := Library{
		DEK:    base64.StdEncoding.EncodeToString(secrets.DEK),
		Series: make(map[string][]Episode),
	}
	key := "library.json.enc"

	return loadData(loadDataInput[Library]{
		ctx:      ctx,
		s3Client: s3Client,
		dType:    library,
		objkey:   key,
		zero:     zero,
		secret:   secrets.KEK,
	})
}

func pushLibrary(ctx context.Context, s3Client *s3.Client, secrets crypto.Secrets, data Library, etag *string) error {
	key := "library.json.enc"

	return pushData(pushDataInput[Library]{
		ctx:      ctx,
		s3Client: s3Client,
		dType:    library,
		objkey:   key,
		data:     data,
		etag:     etag,
		secret:   secrets.KEK,
	})
}
