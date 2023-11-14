package visualizer

import (
	"image"
	BaseLogger "sscs/logger"

	"github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

type MotionDetector struct {
	logger *logrus.Entry

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

func (m *MotionDetector) Start() {
	go m.loop()
}

func (m *MotionDetector) loop() {
	// Initialize gocv structures needed for motion detection.
	mog2 := gocv.NewBackgroundSubtractorMOG2()

	// Loop to read frames and detect motion.
	for {
		select {
		case frame := <-m.frameChan:
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
				img, err := mat.ToImage()
				if err == nil {
					saveToFile(img)
				}
			}

			// Clean up.
			mat.Close()
			fgMask.Close()

		case <-m.stopCh:
			return // Exit the loop when stop signal is received.
		}
	}
}

func (m *MotionDetector) setupLogger() {
	m.logger = BaseLogger.BaseLogger.WithField("package", "motion-detector")
}
