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
	fr        *FaceDetector
	stopCh    chan struct{}
}

func MewCompositeRecognizer(echan EventChannels) *CompositeRecognizer {
	r := &CompositeRecognizer{
		stopCh: make(chan struct{}),
		fr:     NewFaceDetector(EventChannels{FrameIn: echan.FrameIn, RecogOut: echan.RecogOut}),
	}

	r.setupLogger()

	return r
}

func (r *CompositeRecognizer) Start() error {
	// Ensure the recordings directory exists
	cfg, err := conf.ReadConf()

	err = helpers.EnsureDirectoryExists(cfg.Recognizer.ThumbsDir)
	if err != nil {
		r.logger.Errorf("%v", err)
		return err
	}

	r.view()

	return nil
}

func (r *CompositeRecognizer) Stop() error {
	r.fr.Stop()
	return nil
}

func (r *CompositeRecognizer) view() error {
	go r.fr.view()
	return nil
}

func (r *CompositeRecognizer) setupLogger() {
	r.logger = BaseLogger.BaseLogger.WithField("package", "composite-recognizer")
}
