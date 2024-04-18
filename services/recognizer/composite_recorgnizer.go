package recognizer

import (
	"image"

	"github.com/pedrohba1/SSCS/services/conf"
	"github.com/pedrohba1/SSCS/services/helpers"
	BaseLogger "github.com/pedrohba1/SSCS/services/logger"

	"github.com/sirupsen/logrus"
)

type CompositeRecognizer struct {
	logger *logrus.Entry

	frameChan <-chan image.Image
	fr        *HaarDetector
	cr  	*HaarDetector
	stopCh    chan struct{}

}

func NewCompositeRecognizer(echan EventChannels) *CompositeRecognizer {
    r := &CompositeRecognizer{
        stopCh: make(chan struct{}),
    }
	r.frameChan = echan.FrameIn

    // Create the channels for each HaarDetector
    echan.FrameInCopy1 = make(chan image.Image)
    echan.FrameInCopy2 = make(chan image.Image)

    // Initialize each HaarDetector with its respective channel
    r.fr= NewHaarDetector(EventChannels{FrameIn: echan.FrameInCopy1, RecogOut: echan.RecogOut})
    r.cr = NewHaarDetector(EventChannels{FrameIn: echan.FrameInCopy2, RecogOut: echan.RecogOut})

    r.setupLogger()

    // Start the duplicator goroutine
    go r.duplicateFrames(echan)

    return r
}
func (r *CompositeRecognizer) duplicateFrames(echan EventChannels) {
    for {
        select {
        case frame := <-echan.FrameIn:
            // Send the frame to both HaarDetectors
            echan.FrameInCopy1 <- frame
            echan.FrameInCopy2 <- frame
        case <-r.stopCh:
            return
        }
    }
}


func (r *CompositeRecognizer) Start() error {
	// Ensure the recordings directory exists
	cfg, err := conf.ReadConf()

	err = helpers.EnsureDirectoryExists(cfg.Recognizer.ThumbsDir)
	if err != nil {
		r.logger.Errorf("%v", err)
		return err
	}
	go r.view()

	return nil
}

func (r *CompositeRecognizer) Stop() error {
	r.fr.Stop()
	r.cr.Stop()
	return nil
}

func (r *CompositeRecognizer) view() error {
	go r.fr.view()
	go r.cr.view()
	return nil
}

func (r *CompositeRecognizer) setupLogger() {
	r.logger = BaseLogger.BaseLogger.WithField("package", "composite-recognizer")
}
