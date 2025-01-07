package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/OleG2e/collector/internal/config"
	"github.com/OleG2e/collector/pkg/logging"
	"go.uber.org/zap"
)

type FileStorage struct {
	ctx  context.Context
	l    *logging.ZapLogger
	conf *config.ServerConfig
}

func NewFileStorage(ctx context.Context, l *logging.ZapLogger, conf *config.ServerConfig) *FileStorage {
	return &FileStorage{
		ctx:  ctx,
		l:    l,
		conf: conf,
	}
}

func (f *FileStorage) GetStoreType() StoreType {
	return fileStore
}

func (f *FileStorage) store(m *Metrics) error {
	data, err := json.Marshal(m)

	if err != nil {
		return err
	}

	file, fileErr := os.OpenFile(f.conf.FileStoragePath, os.O_RDWR|os.O_CREATE, 0o666)
	defer func(file *os.File) {
		fileCloseErr := file.Close()
		if fileCloseErr != nil && !errors.Is(fileCloseErr, os.ErrClosed) {
			f.l.ErrorCtx(f.ctx, "file close error", zap.Error(fileCloseErr))
		}
	}(file)

	if fileErr != nil {
		return fileErr
	}

	f.l.DebugCtx(f.ctx, "flush storage", zap.String("path", f.conf.FileStoragePath))

	_, err = file.WriteAt(data, 0)

	return err
}

func (f *FileStorage) restore() (*Metrics, error) {
	file, fileErr := os.OpenFile(f.conf.FileStoragePath, os.O_RDWR|os.O_CREATE, 0o666)

	defer func(file *os.File) {
		fileCloseErr := file.Close()
		if fileCloseErr != nil && !errors.Is(fileCloseErr, os.ErrClosed) {
			f.l.ErrorCtx(f.ctx, "file close error", zap.Error(fileCloseErr))
		}
	}(file)

	if fileErr != nil {
		return nil, fileErr
	}

	dec := json.NewDecoder(bufio.NewReader(file))

	lastState := newMetrics()
	if err := dec.Decode(&lastState); err != nil && err != io.EOF {
		f.l.FatalCtx(f.ctx, "restore storage error", zap.Error(err))
	}

	return lastState, nil
}
