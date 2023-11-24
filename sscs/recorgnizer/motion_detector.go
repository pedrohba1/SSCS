package recorgnizer

import (
	"image"
	"sscs/conf"
	"sscs/helpers"
	BaseLogger "sscs/logger"
	"sync"

	"github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

type MotionDetector struct {
	logger *logrus.Entry
	wg     sync.WaitGroup

	frameChan <-chan image.Image
	stopCh    chan struct{}
}

func NewMotionDetector(fchan chan image.Image) *MotionDetector {
	r := &MotionDetector{
		frameChan: fchan,
		stopCh:    make(chan struct{}),
	}
	r.setupLogger()

	return r
}

func (m *MotionDetector) Start() error {
	cfg, _ := conf.ReadConf()
	err := helpers.EnsureDirectoryExists(cfg.Recorgnizer.ThumbsDir)

	if err != nil {
		m.logger.Errorf("%v", err)
		return err
	}
	m.wg.Add(1)
	go m.view()
	return nil
}

func (m *MotionDetector) Stop() error {
	close(m.stopCh) // signal to stop the view
	m.wg.Wait()     // Wait for the recording goroutine to finish
	return nil
}

func (m *MotionDetector) view() error {
	defer m.wg.Done()

	// Initialize gocv structures needed for motion detection.
	mog2 := gocv.NewBackgroundSubtractorMOG2()
	cfg, _ := conf.ReadConf()

	// Loop to read frames and detect motion.
	for {
		select {
		case frame, ok := <-m.frameChan:
			if !ok {
				// channel was closed and drained, handle the closure, perhaps break the view
				break
			}
			if frame == nil {
				m.logger.Info("nil frame received, continuing...")
				continue
			}
			// Convert image.Image to gocv.Mat.
			mat, err := gocv.ImageToMatRGB(frame)
			if err != nil {
				m.logger.Errorf("Error converting image to Mat: %v", err)
				continue
			}

			// Apply the background subtractor to detect motion.
			fgMask := gocv.NewMat()
			mog2.Apply(mat, &fgMask)

			// Check for motion in the foreground mask.
			if gocv.CountNonZero(fgMask) > 0 {
				// Motion detected, save the frame.
				helpers.SaveToFile(frame, cfg.Recorgnizer.ThumbsDir)
			}

			// Clean up.
			mat.Close()
			fgMask.Close()

		case <-m.stopCh:
			m.logger.Info("received stop signal")
			return nil // Exit the view when stop signal is received.
		}
	}
}

func (m *MotionDetector) setupLogger() {
	m.logger = BaseLogger.BaseLogger.WithField("package", "motion-detector")
}
