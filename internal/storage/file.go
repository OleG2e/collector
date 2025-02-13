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
	return FileStoreType
}

func (f *FileStorage) store(m *Metrics) error {
	var tmpFileName string
	err := func() error {
		dir := path.Dir(f.conf.FileStoragePath)
		tmpFile, tmpFileErr := os.CreateTemp(dir, "collector-*.bak")
		if tmpFileErr != nil {
			return tmpFileErr
		}

		tmpFileName = tmpFile.Name()

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

		return nil
	}()

	if err != nil {
		removeErr := os.Remove(tmpFileName)
		if removeErr != nil {
			f.l.WarnCtx(f.ctx, "tmp file remove error", zap.Error(removeErr))
		}
		return err
	}

	err = os.Rename(tmpFileName, f.conf.FileStoragePath)

	return err
}

func (f *FileStorage) restore() (*Metrics, error) {
	file, fileErr := os.Open(f.conf.FileStoragePath)

	defer func(file *os.File) {
		if file == nil {
			f.l.WarnCtx(f.ctx, "restore file doesn't exist")
			return
		}

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
