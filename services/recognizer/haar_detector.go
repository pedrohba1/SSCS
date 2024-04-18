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

// HaarDetector is a type representing
// the face detector component
type HaarDetector struct {
	logger *logrus.Entry
	wg     sync.WaitGroup

	eChans EventChannels
	haarPath string
	thumbsDir  string
	eventName string
	frameLabel string
	stopCh chan struct{}
}

func NewHaarDetector(eChans EventChannels) *HaarDetector {
	r := &HaarDetector{
		eChans: eChans,
		stopCh: make(chan struct{}),
	}
	r.setupLogger()

	return r
}

func (hd *HaarDetector) Start() error {
	// Ensure the recordings directory exists
	cfg, _ := conf.ReadConf()
	hd.haarPath = cfg.Recognizer.HaarPath
	hd.eventName = cfg.Recognizer.EventName
	hd.logger.Info(cfg)
	hd.logger.Info("haar path:", hd.haarPath)

	hd.thumbsDir = cfg.Recognizer.ThumbsDir
	hd.frameLabel = cfg.Recognizer.FrameLabel
	err := helpers.EnsureDirectoryExists(cfg.Recognizer.ThumbsDir)
	if err != nil {
		hd.logger.Errorf("%v", err)
		return err
	}
	hd.wg.Add(1)
	go hd.view()
	return nil
}

func (r *HaarDetector) Stop() error {
	close(r.stopCh) // signal to stop the view
	r.wg.Wait()     // Wait for the recording goroutine to finish
	return nil
}


func (m *HaarDetector) sendRecog(recog RecognizedEvent) error {
	select {
	case m.eChans.RecogOut <- recog:
		return nil
	case <-m.stopCh:
		m.logger.Info("received stop signal")
		return nil
	default:
		m.logger.Info("buffer is full")
		return nil
	}
}

func (r *HaarDetector) view() error {
	defer r.wg.Done()

	// load classifier to recognize faces
	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	if !classifier.Load(r.haarPath) {
		fmt.Printf("Error reading cascade file: %v\n", r.haarPath)
		return fmt.Errorf("couldn't read haar cascading file")
	}
	blue := color.RGBA{0, 0, 255, 0}

	img := gocv.NewMat()
	defer img.Close()

	// Loop to read frames and detect motion.
	for {
		select {
		case frame, ok := <-r.eChans.FrameIn:
			if !ok {
				// channel was closed and drained, handle the closure, perhaps break the view
				break
			}
			if frame == nil {
				r.logger.Info("nil frame received, continuing...")
				continue
			}
			// Convert image.Image to gocv.Mat.
			img, err := gocv.ImageToMatRGB(frame)

			if err != nil {
				r.logger.Errorf("Error converting image to Mat: %v", err)
				continue
			}

			// detect faces
			rects := classifier.DetectMultiScale(img)
			// m.logger.Info("detected faces amount: ", len(rects))

			// draw a rectangle around each face on the original image,
			// along with text identifying as "Human"
			for _, rect := range rects {
				gocv.Rectangle(&img, rect, blue, 3)
				size := gocv.GetTextSize(r.frameLabel, gocv.FontHersheyPlain, 1.2, 2)
				pt := image.Pt(rect.Min.X+(rect.Min.X/2)-(size.X/2), rect.Min.Y-2)
				gocv.PutText(&img, r.frameLabel, pt, gocv.FontHersheyPlain, 1.2, blue, 2)
			}

			helpers.SaveMatToFile(img, r.thumbsDir)

			fname, err:= helpers.SaveMatToFile(img, r.thumbsDir);
			if err != nil {
				r.logger.Errorf("Error saving file: %v", err)
				continue
			}
			r.sendRecog(RecognizedEvent{
				Path: fname,
				Context: r.eventName,
			})
		

		case <-r.stopCh:
			r.logger.Info("received stop signal")
			return nil // Exit the view when stop signal is received.
		}
	}
}

func (m *HaarDetector) setupLogger() {
	m.logger = BaseLogger.BaseLogger.WithField("package", "face-detector")
}
