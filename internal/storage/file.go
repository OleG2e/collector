package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"path"

	"github.com/OleG2e/collector/internal/config"
	"github.com/OleG2e/collector/pkg/logging"
	"go.uber.org/zap"
)

type FileStorage struct {
	ctx  context.Context
	l    *logging.ZapLogger
	conf *config.ServerConfig
}

func NewFileStorage(ctx context.Context, l *logging.ZapLogger, conf *config.ServerConfig) (*FileStorage, error) {
	return &FileStorage{
		ctx:  ctx,
		l:    l,
		conf: conf,
	}, nil
}

func (f *FileStorage) GetStoreType() StoreType {
	return fileStore
}

func (f *FileStorage) store(m *Metrics) error {
	dir := path.Dir(f.conf.FileStoragePath)
	tmpFile, tmpFileErr := os.CreateTemp(dir, "collector-*.bak")
	if tmpFileErr != nil {
		return tmpFileErr
	}

	defer func(tmpFile *os.File) {
		fileCloseErr := tmpFile.Close()
		if fileCloseErr != nil {
			f.l.WarnCtx(f.ctx, "tmp file close error", zap.Error(fileCloseErr))
		}
	}(tmpFile)

	data, err := json.Marshal(m)

	if err != nil {
		return err
	}

	_, err = tmpFile.Write(data)

	if err != nil {
		return err
	}

	err = os.Rename(tmpFile.Name(), f.conf.FileStoragePath)

	return err
}

func (f *FileStorage) restore() (*Metrics, error) {
	file, fileErr := os.Open(f.conf.FileStoragePath)

	defer func(file *os.File) {
		fileCloseErr := file.Close()
		if fileCloseErr != nil {
			f.l.WarnCtx(f.ctx, "file close error", zap.Error(fileCloseErr))
		}
	}(file)

	if file == nil {
		return nil, nil
	}

	if fileErr != nil {
		f.l.WarnCtx(f.ctx, "open DB file error", zap.Error(fileErr))
		return nil, fileErr
	}

	dec := json.NewDecoder(bufio.NewReader(file))

	lastState := newMetrics()
	if err := dec.Decode(&lastState); err != nil && err != io.EOF {
		f.l.FatalCtx(f.ctx, "restore storage error", zap.Error(err))
	}

	return lastState, nil
}

func (f *FileStorage) CloseStorage() error {
	return nil
}
