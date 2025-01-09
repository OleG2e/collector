package storage

import (
	"bufio"
	"context"
	"encoding/json"
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
	tmpFile, tmpFileErr := os.CreateTemp(".", "collector-*.bak")
	if tmpFileErr != nil {
		return tmpFileErr
	}

	defer func(tmpFile *os.File) {
		fileCloseErr := tmpFile.Close()
		if fileCloseErr != nil {
			f.l.WarnCtx(f.ctx, "tmp file close error", zap.Error(fileCloseErr))
		}
	}(tmpFile)

	data, err := json.Marshal(&m)

	if err != nil {
		return err
	}

	_, err = tmpFile.Write(data)

	if err != nil {
		return err
	}

	err = os.Rename(tmpFile.Name(), f.conf.FileStoragePath)

	f.l.InfoCtx(f.ctx, "flush storage")

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
