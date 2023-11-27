package recognizer

import (
	"fmt"
	"image"
	"image/color"
	"sync"

	"github.com/pedrohba1/SSCS/services/conf"
	"github.com/pedrohba1/SSCS/services/helpers"
	BaseLogger "github.com/pedrohba1/SSCS/services/logger"

	"github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

// FaceDetector is a type representing
// the face detector component
type FaceDetector struct {
	logger *logrus.Entry
	wg     sync.WaitGroup

	eChans EventChannels
	stopCh chan struct{}
}

func NewFaceDetector(eChans EventChannels) *FaceDetector {
	r := &FaceDetector{
		eChans: eChans,
		stopCh: make(chan struct{}),
	}
	r.setupLogger()

	return r
}

func (fd *FaceDetector) Start() error {
	// Ensure the recordings directory exists
	cfg, _ := conf.ReadConf()
	err := helpers.EnsureDirectoryExists(cfg.Recognizer.ThumbsDir)
	if err != nil {
		fd.logger.Errorf("%v", err)
		return err
	}
	fd.wg.Add(1)
	go fd.view()
	return nil
}

func (m *FaceDetector) Stop() error {
	close(m.stopCh) // signal to stop the view
	m.wg.Wait()     // Wait for the recording goroutine to finish
	return nil
}

func (m *FaceDetector) view() error {
	defer m.wg.Done()

	// load classifier to recognize faces
	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	cfg, _ := conf.ReadConf()
	if !classifier.Load(cfg.Recognizer.FaceHaarPath) {
		fmt.Printf("Error reading cascade file: %v\n", cfg.Recognizer.FaceHaarPath)
		return fmt.Errorf("couldn't read haar cascading file")
	}
	blue := color.RGBA{0, 0, 255, 0}

	// Loop to read frames and detect motion.
	for {
		select {
		case frame, ok := <-m.eChans.FrameIn:
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

			// detect faces
			rects := classifier.DetectMultiScale(mat)
			// m.logger.Info("detected faces amount: ", len(rects))

			// draw a rectangle around each face on the original image,
			// along with text identifying as "Human"
			for _, r := range rects {
				gocv.Rectangle(&mat, r, blue, 3)
				size := gocv.GetTextSize("Human", gocv.FontHersheyPlain, 1.2, 2)
				pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
				gocv.PutText(&mat, "Human", pt, gocv.FontHersheyPlain, 1.2, blue, 2)
			}

			helpers.SaveMatToFile(mat, cfg.Recognizer.ThumbsDir)
			mat.Close()

		case <-m.stopCh:
			m.logger.Info("received stop signal")
			return nil // Exit the view when stop signal is received.
		}
	}
}

func (m *FaceDetector) setupLogger() {
	m.logger = BaseLogger.BaseLogger.WithField("package", "face-detector")
}
