package recognizer

import (
	"image"
	"image/color"
	"sync"

	"github.com/pedrohba1/SSCS/services/conf"
	"github.com/pedrohba1/SSCS/services/helpers"
	BaseLogger "github.com/pedrohba1/SSCS/services/logger"

	"github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

type MotionDetector struct {
	logger *logrus.Entry
	wg     sync.WaitGroup

	MinimumArea int
	thumbsDir string
	eChans      EventChannels
	stopCh      chan struct{}
}

func NewMotionDetector(eChans EventChannels) *MotionDetector {

	cfg, _ := conf.ReadConf()
	r := &MotionDetector{
		eChans:      eChans,
		MinimumArea: 3000,
		thumbsDir: cfg.Recognizer.ThumbsDir,
		stopCh: make(chan struct{}),
	}
	r.setupLogger()

	return r
}

func (m *MotionDetector) Start() error {
	m.logger.Info("starting motion detector...")
	err := helpers.EnsureDirectoryExists(m.thumbsDir)

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
	img := gocv.NewMat()
	defer img.Close()

	imgDelta := gocv.NewMat()
	defer imgDelta.Close()

	imgThresh := gocv.NewMat()
	defer imgThresh.Close()

	mog2 := gocv.NewBackgroundSubtractorMOG2()
	defer mog2.Close()

	status := "Ready"
	statusColor := color.RGBA{0, 255, 0, 0}

	// Loop to read frames and detect motion.

	for {
		select {
		case frame, ok := <-m.eChans.FrameIn:
			if !ok {
				// channel was closed and drained, handle the closure, perhaps break the view
				break
			}
			if frame == nil {
				m.logger.Warn("nil frame received, continuing...")
				continue
			}
			// Convert image.Image to gocv.Mat.
			img, err := gocv.ImageToMatRGB(frame)
			if err != nil {
				m.logger.Errorf("Error converting image to Mat: %v", err)
				continue
			}

			// first phase of cleaning up image, obtain foreground only
			mog2.Apply(img, &imgDelta)

			// remaining cleanup of the image to use for finding contours.
			// first use threshold
			gocv.Threshold(imgDelta, &imgThresh, 25, 255, gocv.ThresholdBinary)

			// then dilate
			kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
			gocv.Dilate(imgThresh, &imgThresh, kernel)
			kernel.Close()

			// now find contours
			contours := gocv.FindContours(imgThresh, gocv.RetrievalExternal, gocv.ChainApproxSimple)
			for i := 0; i < contours.Size(); i++ {
				area := gocv.ContourArea(contours.At(i))
				if area < float64(m.MinimumArea) {
					continue
				}

				status = "Motion detected"
				statusColor = color.RGBA{255, 0, 0, 0}
				gocv.DrawContours(&img, contours, i, statusColor, 2)

				rect := gocv.BoundingRect(contours.At(i))
				gocv.Rectangle(&img, rect, color.RGBA{0, 0, 255, 0}, 2)
			}

			contours.Close()

			gocv.PutText(&img, status, image.Pt(10, 20), gocv.FontHersheyPlain, 1.2, statusColor, 2)
			helpers.SaveMatToFile(img, m.thumbsDir)

		case <-m.stopCh:
			m.logger.Info("received stop signal")
			return nil // Exit the view when stop signal is received.
		}
	}
}

func (m *MotionDetector) setupLogger() {
	m.logger = BaseLogger.BaseLogger.WithField("package", "motion-detector")
}
