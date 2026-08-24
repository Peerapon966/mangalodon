package task

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Peerapon966/blackbox/scraper/internal/apperr"
	"github.com/Peerapon966/blackbox/scraper/internal/crawler"
	"github.com/Peerapon966/blackbox/scraper/internal/crypto"
	"github.com/Peerapon966/blackbox/scraper/internal/s3"
	"github.com/Peerapon966/blackbox/scraper/internal/utils"
)

type Task struct {
	S3Client   *s3.Client
	Secrets    crypto.Secrets
	Ctx        context.Context
	Crawler    crawler.Crawler
	ID         string     // autofilled
	Status     TaskStatus // autofilled
	Registered bool       // autofilled
}

type TaskStatus string

const (
	Unknown    TaskStatus = ""
	Pending    TaskStatus = "PENDING"
	InProgress TaskStatus = "IN_PROGRESS"
	Completed  TaskStatus = "COMPLETED"
	Failed     TaskStatus = "FAILED"
)

type ActiveTaskItem struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Title     string `json:"title"`
	Episode   int    `json:"episode"`
	CreatedAt string `json:"createdAt"`
}

type ActiveTaskIndex struct {
	Tasks []ActiveTaskItem `json:"tasks"`
}

type HistoryTaskItem struct {
	ID        string               `json:"id"`
	URL       string               `json:"url"`
	Title     string               `json:"title"`
	Episode   int                  `json:"episode"`
	Status    TaskStatus           `json:"status"`
	Error     *apperr.ScraperError `json:"error"`
	CreatedAt string               `json:"createdAt"`
}

type HistoryTaskIndex struct {
	Tasks []HistoryTaskItem `json:"tasks"`
}

type TaskProgress struct {
	Message TaskStatus `json:"message"`
	Current int        `json:"current"`
	Total   int        `json:"total"`
}

type TaskState struct {
	Status   TaskStatus          `json:"status"`
	Progress TaskProgress        `json:"progress"`
	Error    apperr.ScraperError `json:"error"`
}

type Episode struct {
	ID      string `json:"id"`
	Episode int    `json:"episode"`
	Pages   int    `json:"pages"`
	EXT     string `json:"ext"`
}

type Library struct {
	DEK    string               `json:"dek"`
	Series map[string][]Episode `json:"series"`
}

type DeregisterTaskInput struct {
	S3Client *s3.Client
	Secrets  crypto.Secrets
	Ctx      context.Context
	Status   TaskStatus
	TaskID   string
	URL      string
	Title    string
	Episode  int
	Error    *apperr.ScraperError
}

type NewTaskInput struct {
	Crawler  crawler.Crawler
	S3Client *s3.Client
	Secrets  crypto.Secrets
}

func New(ctx context.Context, params NewTaskInput) (*Task, error) {
	// init crawler first to load title in case empty
	err := params.Crawler.Initialize()
	if err != nil {
		return nil, err
	}

	t := Task{
		S3Client: params.S3Client,
		Secrets:  params.Secrets,
		Ctx:      ctx,
		Crawler:  params.Crawler,
		Status:   Pending,
	}
	input := fmt.Sprintf("%s_%d", params.Crawler.GetTitle(), params.Crawler.GetEpisode())
	hash := sha256.Sum256([]byte(input))
	t.ID = hex.EncodeToString(hash[:])
	err = utils.RetryWithNoResponse(t.registerActiveTask, utils.RetryConfig{
		Attempts: 5,
		Delay:    3 * time.Second,
		FatalErrors: []apperr.ErrorCode{
			apperr.NoSuchKey,
			apperr.TaskAlreadyExists,
		},
	})
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// 1. add failed task to data/history/YYYY-MM.json.enc
// 2. deregister taskID from data/tasks.json.enc
// 3. delete data/tasks/<taskID>.json.enc
func DeregisterTask(params DeregisterTaskInput) error {
	if err := utils.RetryWithNoResponse(func() error {
		return registerHistoryTask(params)
	}, utils.RetryConfig{
		Attempts: 3,
		Delay:    5 * time.Second,
	}); err != nil {
		return err
	}

	err := utils.RetryWithNoResponse(func() error {
		return deregisterActiveTask(params)
	}, utils.RetryConfig{
		Attempts: 3,
		Delay:    5 * time.Second,
	})

	return err
}

func (t *Task) Start() error {
	slog.Info(fmt.Sprintf("Starting scraping task for task ID: %s", t.ID))
	retryCfg := utils.RetryConfig{
		Attempts: 3,
		Delay:    3 * time.Second,
		FatalErrors: []apperr.ErrorCode{
			apperr.TaskAlreadyExists,
		},
	}
	err := utils.RetryWithNoResponse(func() error {
		return t.updateTaskState(TaskState{
			Status: InProgress,
			Progress: TaskProgress{
				Current: 0,
				Total:   max(t.Crawler.GetPageCount(), 1),
				Message: "Scrape task registered. Starting...",
			},
		})
	}, retryCfg)
	if err != nil {
		return err
	}

	err = t.Crawler.LoadPageLinks()
	if err != nil {
		return err
	}

	var completedPages int32 = 0
	var updateMutex sync.Mutex

	// a function that encrypts and upload images to an S3 bucket
	// pass this function as a callback to Crawler.Crawl to call after fetching a page
	imgProcessor := func(img []byte, page int) error {
		encImage, err := crypto.EncryptBlob(img, t.Secrets.DEK)
		if err != nil {
			return err
		}

		err = t.S3Client.UploadObject(t.Ctx, &s3.UploadObjectInput{
			ObjectKey: fmt.Sprintf("data/blobs/%s/%s.%s.enc", t.ID, utils.If(page == 0, "cover", strconv.Itoa(page)), t.Crawler.GetImageExt()),
			Data:      encImage,
		})
		if page <= 0 || err != nil {
			return err
		}

		currentCompleted := atomic.AddInt32(&completedPages, 1)
		if (currentCompleted%5 == 0 || currentCompleted == int32(t.Crawler.GetPageCount())) &&
			updateMutex.TryLock() {
			go func(current int32) {
				defer updateMutex.Unlock()

				_ = utils.RetryWithNoResponse(func() error {
					return t.updateTaskState(TaskState{
						Status: InProgress,
						Progress: TaskProgress{
							Current: int(current),
							Total:   t.Crawler.GetPageCount(),
							Message: "Crawling images...",
						},
					})
				}, retryCfg)
			}(currentCompleted)
		}

		return err
	}

	slog.Info("Starting image crawl.")
	err = t.Crawler.Crawl(t.Ctx, imgProcessor)
	if err != nil {
		return err
	}
	slog.Info(fmt.Sprintf("Successfully crawled %d images.", t.Crawler.GetPageCount()+1))

	err = utils.RetryWithNoResponse(func() error {
		return t.updateTaskState(TaskState{
			Status: InProgress,
			Progress: TaskProgress{
				Current: t.Crawler.GetPageCount() + 1,
				Total:   t.Crawler.GetPageCount() + 1,
				Message: "Scraping completed, registering episode...",
			},
		})
	}, retryCfg)
	if err != nil {
		return err
	}

	err = utils.RetryWithNoResponse(t.registerEpisode, utils.RetryConfig{
		Attempts: 3,
		Delay:    5 * time.Second,
	})
	if err != nil {
		return err
	}
	slog.Info("Successfully registered episode.")

	err = utils.RetryWithNoResponse(func() error {
		return t.updateTaskState(TaskState{
			Status: InProgress,
			Progress: TaskProgress{
				Current: (t.Crawler.GetPageCount() + 1) * 2,
				Total:   (t.Crawler.GetPageCount() + 1) * 2,
				Message: "Episode registered, cleaning up...",
			},
		})
	}, retryCfg)
	if err != nil {
		return err
	}

	err = utils.RetryWithNoResponse(func() error {
		return t.updateTaskState(TaskState{
			Status: Completed,
			Progress: TaskProgress{
				Current: (t.Crawler.GetPageCount() + 1) * 2,
				Total:   (t.Crawler.GetPageCount() + 1) * 2,
				Message: "Scrape task completed.",
			},
		})
	}, retryCfg)
	if err != nil {
		return err
	}

	err = DeregisterTask(DeregisterTaskInput{
		S3Client: t.S3Client,
		Secrets:  t.Secrets,
		Ctx:      t.Ctx,
		Status:   Completed,
		TaskID:   t.ID,
		URL:      t.Crawler.GetURL(),
		Title:    t.Crawler.GetTitle(),
		Episode:  t.Crawler.GetEpisode(),
	})
	slog.Info("Successfully deregistered task.")

	return err
}

// Generate and upload TaskState object to S3 data/tasks/<taskId>.json.enc
// Generate and register new ActiveTaskItem object to S3 data/tasks.json.enc
func (t *Task) registerActiveTask() error {
	initialTaskState := TaskState{
		Status: Pending,
		Progress: TaskProgress{
			Current: 0,
			Total:   max(t.Crawler.GetPageCount(), 1),
			Message: "Registering scrape task...",
		},
	}
	err := utils.RetryWithNoResponse(func() error {
		return t.updateTaskState(initialTaskState)
	},
		utils.RetryConfig{
			Attempts: 3,
			Delay:    3 * time.Second,
			FatalErrors: []apperr.ErrorCode{
				apperr.TaskAlreadyExists,
			},
		})
	if err != nil {
		return err
	}

	activeTaskIndex, etag, err := loadActiveTaskIndex(t.Ctx, t.S3Client, t.Secrets)
	if err != nil {
		return err
	}

	activeTaskIndex.Tasks = append(activeTaskIndex.Tasks, ActiveTaskItem{
		ID:        t.ID,
		URL:       t.Crawler.GetURL(),
		Title:     t.Crawler.GetTitle(),
		Episode:   t.Crawler.GetEpisode(),
		CreatedAt: time.Now().Format(time.RFC3339),
	})

	err = pushActiveTaskIndex(t.Ctx, t.S3Client, t.Secrets, activeTaskIndex, etag)
	if err != nil {
		return err
	}

	t.Registered = true

	return nil
}

func (t *Task) registerEpisode() error {
	library, etag, err := loadLibrary(t.Ctx, t.S3Client, t.Secrets)
	if err != nil {
		return err
	}

	episode := Episode{
		ID:      t.ID,
		Episode: t.Crawler.GetEpisode(),
		Pages:   t.Crawler.GetPageCount(),
		EXT:     t.Crawler.GetImageExt(),
	}
	if library.Series == nil {
		library.Series = make(map[string][]Episode)
	}

	epExists := false
	if series, exists := library.Series[t.Crawler.GetTitle()]; exists && series != nil {
		for i, ep := range series {
			if ep.Episode == t.Crawler.GetEpisode() {
				library.Series[t.Crawler.GetTitle()][i].ID = t.ID
				library.Series[t.Crawler.GetTitle()][i].Pages = t.Crawler.GetPageCount()
				library.Series[t.Crawler.GetTitle()][i].EXT = t.Crawler.GetImageExt()
				epExists = true
				break
			}
		}
	}

	if !epExists {
		library.Series[t.Crawler.GetTitle()] = append(library.Series[t.Crawler.GetTitle()], episode)
	}
	err = pushLibrary(t.Ctx, t.S3Client, t.Secrets, library, etag)

	return err
}

func (t *Task) updateTaskState(taskState TaskState) error {
	var etag *string
	var err error
	if t.Registered {
		_, etag, err = loadActiveTaskState(t.Ctx, t.S3Client, t.Secrets, t.ID)
		if err != nil {
			return err
		}
	}

	mode := utils.If(t.Registered, s3.IfMatch, s3.IfNoneMatch)
	err = pushActiveTaskState(t.Ctx, t.S3Client, t.Secrets, t.ID, taskState, etag, mode)
	if err != nil {
		var scraperErr *apperr.ScraperError
		if errors.As(err, &scraperErr) && scraperErr.ErrorCode() == apperr.IfNoneMatchPreconditionFailed {
			return &apperr.ScraperError{
				Code:    apperr.TaskAlreadyExists,
				Message: fmt.Sprintf("Active task with ID '%s' already exists. Dropped the request.", t.ID),
			}
		} else if errors.As(err, &scraperErr) && scraperErr.ErrorCode() == apperr.IfMatchPreconditionFailed {
			return &apperr.ScraperError{
				Code:    apperr.IfMatchPreconditionFailed,
				Message: fmt.Sprintf("Race condition detected when trying to update task state for active task '%s'", t.ID),
			}
		}
	}

	return err
}

func registerHistoryTask(params DeregisterTaskInput) error {
	loc := time.FixedZone("UTC+7", 7*60*60)
	now := time.Now().In(loc)
	historyTaskItem := HistoryTaskItem{
		ID:        params.TaskID,
		URL:       params.URL,
		Title:     params.Title,
		Episode:   params.Episode,
		Status:    params.Status,
		Error:     params.Error,
		CreatedAt: now.String(),
	}

	historyTaskIndex, etag, err := loadHistoryTaskIndex(params.Ctx, params.S3Client, params.Secrets, now.Year(), int(now.Month()))
	if err != nil {
		return err
	}

	historyTaskIndex.Tasks = append(historyTaskIndex.Tasks, historyTaskItem)
	err = pushHistoryTaskIndex(params.Ctx, params.S3Client, params.Secrets, historyTaskIndex, etag, now.Year(), int(now.Month()))
	if err != nil {
		return err
	}

	return nil
}

func deregisterActiveTask(params DeregisterTaskInput) error {
	activeTaskIndex, etag, err := loadActiveTaskIndex(params.Ctx, params.S3Client, params.Secrets)
	if err != nil {
		return err
	}

	for i, task := range activeTaskIndex.Tasks {
		if task.ID == params.TaskID {
			activeTaskIndex.Tasks[i] = activeTaskIndex.Tasks[len(activeTaskIndex.Tasks)-1]
			activeTaskIndex.Tasks = activeTaskIndex.Tasks[:len(activeTaskIndex.Tasks)-1]
			break
		}

		if i == len(activeTaskIndex.Tasks)-1 {
			slog.Warn("Couldn't deregister active task from task index, no such task ID found.",
				slog.String("task_id", params.TaskID),
			)
		}
	}

	err = pushActiveTaskIndex(params.Ctx, params.S3Client, params.Secrets, activeTaskIndex, etag)
	if err != nil {
		return err
	}

	slog.Info("deleting active task state",
		slog.String("task_id", params.TaskID),
	)
	err = params.S3Client.RemoveObject(params.Ctx, fmt.Sprintf("data/tasks/status/%s.json.enc", params.TaskID))

	return err
}
