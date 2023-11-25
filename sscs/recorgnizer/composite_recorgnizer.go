package recorgnizer

import (
	"image"

	"github.com/pedrohba1/SSCS/sscs/conf"
	"github.com/pedrohba1/SSCS/sscs/helpers"
	BaseLogger "github.com/pedrohba1/SSCS/sscs/logger"

	"github.com/sirupsen/logrus"
)

type CompositeRecorgnizer struct {
	logger *logrus.Entry

	frameChan <-chan image.Image
	fr        *FaceDetector
	stopCh    chan struct{}
}

func MewCompositeRecorgnizer(fchan chan image.Image) *CompositeRecorgnizer {
	r := &CompositeRecorgnizer{
		frameChan: fchan,
		stopCh:    make(chan struct{}),
		fr:        NewFaceDetector(fchan),
	}

	r.setupLogger()

	return r
}

func (r *CompositeRecorgnizer) Start() error {
	// Ensure the recordings directory exists
	cfg, err := conf.ReadConf()

	err = helpers.EnsureDirectoryExists(cfg.Recorgnizer.ThumbsDir)
	if err != nil {
		r.logger.Errorf("%v", err)
		return err
	}

	r.view()

	return nil
}

func (r *CompositeRecorgnizer) Stop() error {
	r.fr.Stop()
	return nil
}

func (r *CompositeRecorgnizer) view() error {
	go r.fr.view()
	return nil
}

func (r *CompositeRecorgnizer) setupLogger() {
	r.logger = BaseLogger.BaseLogger.WithField("package", "composite-recorgnizer")
}
