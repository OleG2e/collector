package store

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"sync"

	"collector/internal/config"
	"collector/internal/core/domain"
)

type FileStorage struct {
	logger  *slog.Logger
	conf    *config.ServerConfig
	metrics *domain.Metrics
	mx      *sync.RWMutex
}

func NewFileStorage(
	logger *slog.Logger,
	conf *config.ServerConfig,
	metrics *domain.Metrics,
) (*FileStorage, error) {
	if err := pingFS(logger); err != nil {
		return nil, fmt.Errorf("(file) ping filesystem error: %w", err)
	}

	return &FileStorage{
		logger:  logger,
		conf:    conf,
		metrics: metrics,
		mx:      new(sync.RWMutex),
	}, nil
}

func (f *FileStorage) GetStoreType() Type {
	return FileStoreType
}

func (f *FileStorage) Save(ctx context.Context) error {
	var tmpFileName string

	err := func() error {
		dir := path.Dir(f.conf.FileStoragePath)

		tmpFile, tmpFileErr := os.CreateTemp(dir, "collector-*.bak")
		if tmpFileErr != nil {
			return fmt.Errorf("(file) create tmp file error: %w", tmpFileErr)
		}

		tmpFileName = tmpFile.Name()

		defer func(tmpFile *os.File) {
			fileCloseErr := tmpFile.Close()
			if fileCloseErr != nil {
				f.logger.WarnContext(
					ctx,
					"(file) tmp file close error",
					slog.Any("error", fileCloseErr),
				)
			}
		}(tmpFile)

		data, marshErr := json.Marshal(f.GetMetrics())
		if marshErr != nil {
			return fmt.Errorf("(file) marshall metrics data error: %w", marshErr)
		}

		_, tmpFileWriteErr := tmpFile.Write(data)
		if tmpFileWriteErr != nil {
			return fmt.Errorf("(file) tmp file write error: %w", tmpFileWriteErr)
		}

		return nil
	}()
	if err != nil {
		removeErr := os.Remove(tmpFileName)
		if removeErr != nil {
			return fmt.Errorf("(file) tmp file remove error: %w", removeErr)
		}

		return nil
	}

	err = os.Rename(tmpFileName, f.conf.FileStoragePath)
	if err != nil {
		return fmt.Errorf("(file) tmp file rename error: %w", err)
	}

	return nil
}

func (f *FileStorage) Restore(ctx context.Context) error {
	file, fileErr := os.Open(f.conf.FileStoragePath)

	defer func(file *os.File) {
		if file == nil {
			return
		}

		fileCloseErr := file.Close()
		if fileCloseErr != nil {
			f.logger.WarnContext(ctx, "(file) file close error", slog.Any("error", fileCloseErr))
		}
	}(file)

	if file == nil {
		f.logger.WarnContext(ctx, "(file) restore file doesn't exist")

		return nil
	}

	if fileErr != nil {
		return fmt.Errorf("(file) open restore file error: %w", fileErr)
	}

	dec := json.NewDecoder(bufio.NewReader(file))
	lastState := domain.NewMetrics()

	if err := dec.Decode(&lastState); err != nil && err != io.EOF {
		return fmt.Errorf("(file) restore storage error: %w", err)
	}

	f.SetMetrics(lastState)

	return nil
}

func (f *FileStorage) GetMetrics() *domain.Metrics {
	f.mx.RLock()
	defer f.mx.RUnlock()

	return f.metrics
}

func (f *FileStorage) SetMetrics(metrics *domain.Metrics) {
	f.mx.Lock()
	defer f.mx.Unlock()

	f.metrics = metrics
}

func (f *FileStorage) Close() error {
	return nil
}

func pingFS(logger *slog.Logger) error {
	tmpFile, tmpFileErr := os.CreateTemp("/tmp", "collector-*.tmp")

	defer func(tmpFile *os.File) {
		fileCloseErr := tmpFile.Close()
		ctx := context.Background()

		if fileCloseErr != nil {
			logger.WarnContext(
				ctx,
				"(file) ping tmp file close error",
				slog.Any("error", fileCloseErr),
			)

			return
		}

		removeErr := os.Remove(tmpFile.Name())
		if removeErr != nil {
			logger.WarnContext(
				ctx,
				"(file) ping tmp file remove error",
				slog.Any("error", removeErr),
			)

			return
		}
	}(tmpFile)

	if tmpFileErr != nil {
		return fmt.Errorf("(file) ping tmp file error: %w", tmpFileErr)
	}

	return nil
}
