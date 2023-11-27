package storer

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/pedrohba1/SSCS/services/conf"
	"github.com/pedrohba1/SSCS/services/helpers"
	BaseLogger "github.com/pedrohba1/SSCS/services/logger"
	"github.com/sirupsen/logrus"
)

type OSStorer struct {
	sizeLimit   int
	checkPeriod int
	folderPath  string
	logger      *logrus.Entry

	wg             sync.WaitGroup
	CleanEventChan chan<- CleanEvent

	stopCh chan struct{}
}

func NewOSStorer() *OSStorer {

	cfg, _ := conf.ReadConf()

	s := &OSStorer{
		sizeLimit:   cfg.Storer.SizeLimit,
		checkPeriod: cfg.Storer.CheckPeriod,
		folderPath:  cfg.Recorder.RecordingsDir,
		stopCh:      make(chan struct{}),
	}
	s.setupLogger()

	return s
}

func (s *OSStorer) setupLogger() {
	s.logger = BaseLogger.BaseLogger.WithField("package", "Storer")
}

func (s *OSStorer) Start() error {
	s.logger.Info("starting cleaning service...")
	err := helpers.EnsureDirectoryExists(s.folderPath)

	if err != nil {
		s.logger.Errorf("%v", err)
		return err
	}

	s.wg.Add(1)
	go s.monitor()
	return nil
}

func (r *OSStorer) Stop() error {
	close(r.stopCh)
	r.wg.Wait()
	return nil
}

func (s *OSStorer) monitor() error {
	defer s.wg.Done()

	ticker := time.NewTicker(time.Duration(s.checkPeriod) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.logger.Info("cleaning storage limit...")
			err := s.checkAndCleanFolder()
			if err != nil {
				s.logger.Error("Error: ", err)
			}

		case <-s.stopCh:
			s.logger.Info("received stop signal")
			return nil
		}
	}
}

func (s *OSStorer) checkAndCleanFolder() error {

	entries, err := os.ReadDir(s.folderPath)
	if err != nil {
		return err

	}
	infos := make([]fs.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return err
		}
		infos = append(infos, info)
	}

	sort.Sort(helpers.ByModTime(infos))

	// Calculate the total size
	var totalSize int64 = 0

	for _, info := range infos {
		totalSize += info.Size()
	}
	s.logger.Infof("Folder size before deletion: %.2f MB ", float64(totalSize)/1024/1024)

	// Delete the oldest files if the total size exceeds the limit
	var deletedSize int64 = 0
	for _, info := range infos {
		// when the limit is reached, break the loop
		if totalSize <= int64(s.sizeLimit) {
			break
		}

		oldestFilePath := filepath.Join(s.folderPath, info.Name())

		err := os.Remove(oldestFilePath)
		if err != nil {
			s.logger.Error("Error removing file: ", err)
			continue
		}
		deletedSize += info.Size()
		totalSize -= info.Size()
		s.CleanEventChan <- CleanEvent{
			filename: info.Name(),
			fileSize: int(info.Size()),
		}
	}

	s.logger.Infof("deleted files size: %.2f MB ", float64(deletedSize)/1024/1024)

	return nil
}
