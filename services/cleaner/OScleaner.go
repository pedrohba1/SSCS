package cleaner

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

type OSCleaner struct {
	sizeLimit   int
	checkPeriod int
	folderPath  string
	logger      *logrus.Entry

	wg             sync.WaitGroup
	CleanEventChan chan<- CleanEvent

	stopCh chan struct{}
}

func NewOSCleaner() *OSCleaner {

	cfg, _ := conf.ReadConf()

	c := &OSCleaner{
		sizeLimit:   cfg.Cleaner.SizeLimit,
		checkPeriod: cfg.Cleaner.CheckPeriod,
		folderPath:  cfg.Recorder.RecordingsDir,
		stopCh:      make(chan struct{}),
	}
	c.setupLogger()

	return c
}

func (c *OSCleaner) setupLogger() {
	c.logger = BaseLogger.BaseLogger.WithField("package", "cleaner")
}

func (c *OSCleaner) Start() error {
	c.logger.Info("starting cleaning service...")
	err := helpers.EnsureDirectoryExists(c.folderPath)

	if err != nil {
		c.logger.Errorf("%v", err)
		return err
	}

	c.wg.Add(1)
	go c.listen()
	return nil
}

func (r *OSCleaner) Stop() error {
	close(r.stopCh)
	r.wg.Wait()
	return nil
}

func (c *OSCleaner) listen() error {
	defer c.wg.Done()

	ticker := time.NewTicker(time.Duration(c.checkPeriod) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.logger.Info("cleaning storage limit...")
			err := c.checkAndCleanFolder()
			if err != nil {
				c.logger.Error("Error: ", err)
			}

		case <-c.stopCh:
			c.logger.Info("received stop signal")
			return nil
		}
	}
}

func (c *OSCleaner) checkAndCleanFolder() error {

	entries, err := os.ReadDir(c.folderPath)
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
	c.logger.Infof("Folder size before deletion: %.2f MB ", float64(totalSize)/1024/1024)

	// Delete the oldest files if the total size exceeds the limit
	var deletedSize int64 = 0
	for _, info := range infos {
		// when the limit is reached, break the loop
		if totalSize <= int64(c.sizeLimit) {
			break
		}

		oldestFilePath := filepath.Join(c.folderPath, info.Name())

		err := os.Remove(oldestFilePath)
		if err != nil {
			c.logger.Error("Error removing file: ", err)
			continue
		}
		deletedSize += info.Size()
		totalSize -= info.Size()
		c.CleanEventChan <- CleanEvent{
			filename: info.Name(),
			fileSize: int(info.Size()),
		}
	}

	c.logger.Infof("deleted files size: %.2f MB ", float64(deletedSize)/1024/1024)

	return nil
}
