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
	logger *logrus.Entry

	wg sync.WaitGroup

	eChans EventChannels
	cfg    Config
	stopCh chan struct{}
}

// creates a new Storer. Notice that
// if no event channel
func NewOSStorer(eChans EventChannels) *OSStorer {

	cfg, _ := conf.ReadConf()

	s := &OSStorer{
		cfg: Config{
			sizeLimit:   cfg.Storer.SizeLimit,
			checkPeriod: cfg.Storer.CheckPeriod,
			folderPath:  cfg.Recorder.RecordingsDir,
			backupPath:  cfg.Storer.BackupPath,
		},
		eChans: eChans,
		stopCh: make(chan struct{}),
	}
	s.setupLogger()

	return s
}

func (s *OSStorer) setupLogger() {
	s.logger = BaseLogger.BaseLogger.WithField("package", "storer")
}

func (s *OSStorer) Start() error {
	s.logger.Info("starting storer service...")

	if s.cfg.backupPath == "" {
		s.logger.Warn("backupPath is not defined in configuration. Files will be erased by default")
	} else {
		err := helpers.EnsureDirectoryExists(s.cfg.backupPath)
		if err != nil {
			s.logger.Errorf("failed to ensure that backup directory exists: %v", err)
		}
	}
	err := helpers.EnsureDirectoryExists(s.cfg.folderPath)

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

	ticker := time.NewTicker(time.Duration(s.cfg.checkPeriod) * time.Second)
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

	entries, err := os.ReadDir(s.cfg.folderPath)
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
		if totalSize <= int64(s.cfg.sizeLimit) {
			break
		}

		oldestFilePath := filepath.Join(s.cfg.folderPath, info.Name())

		// If backupPath is defined, move the file there, otherwise remove the file.
		if s.cfg.backupPath != "" {
			backupFilePath := filepath.Join(s.cfg.backupPath, info.Name())
			err := os.Rename(oldestFilePath, backupFilePath)
			if err != nil {
				s.logger.Error("Error moving file to backup directory: ", err)
				continue
			}
			s.logger.Infof("File moved to backup directory: %s", backupFilePath)
		} else {
			err := os.Remove(oldestFilePath)
			if err != nil {
				s.logger.Error("Error removing file: ", err)
				continue
			}
			s.logger.Infof("File removed: %s", oldestFilePath)
		}

		deletedSize += info.Size()
		totalSize -= info.Size()
		s.eChans.CleanOut <- CleanedEvent{
			filename:   info.Name(),
			fileSize:   int(info.Size()),
			fileStatus: FileErased,
		}
	}

	s.logger.Infof("deleted files size: %.2f MB ", float64(deletedSize)/1024/1024)

	return nil
}

// OpenFiles takes a slice of filenames and attempts to open each one.
// It returns a slice of *os.File and any error encountered.
func (s *OSStorer) OpenFiles(filenames []string) ([]*os.File, error) {
	var files []*os.File
	for _, filename := range filenames {
		file, err := os.Open(filename) // For read access.
		if err != nil {
			// Close all opened files before returning the error
			for _, f := range files {
				f.Close()
			}
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}
