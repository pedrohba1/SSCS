package recorder

import (
	"image"
	"sync"

	"github.com/pedrohba1/SSCS/services/conf"
	"github.com/pedrohba1/SSCS/services/helpers"
	BaseLogger "github.com/pedrohba1/SSCS/services/logger"

	"github.com/aler9/gortsplib/pkg/h264"
	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/bluenviron/gortsplib/v4/pkg/format/rtph264"
	"github.com/pion/rtp"
	"github.com/sirupsen/logrus"
)

type RTSP_H264Recorder struct {
	rtspURL string
	client  *gortsplib.Client
	logger  *logrus.Entry

	wg sync.WaitGroup

	eChans EventChannels
	stopCh chan struct{}
}

// This code requires the FFmpeg libraries, that can be installed with this command:
// apt install -y libavformat-dev libswscale-dev gcc pkg-config
func NewRTSP_H264Recorder(rtspURL string, eChans EventChannels) *RTSP_H264Recorder {
	r := &RTSP_H264Recorder{
		rtspURL: rtspURL,
		eChans:  eChans,
		stopCh:  make(chan struct{}),
	}
	r.setupLogger()

	return r
}

func (r *RTSP_H264Recorder) setupLogger() {
	r.logger = BaseLogger.BaseLogger.WithField("package", "recorder")
}

func (r *RTSP_H264Recorder) Start() error {
	u, err := base.ParseURL(r.rtspURL)

	r.client = &gortsplib.Client{}

	// connect to the server
	err = r.client.Start(u.Scheme, u.Host)
	if err != nil {
		r.logger.Error("failed to start RTSP client: %w", err)
		return err
	}

	// Ensure the recordings directory exists
	cfg, _ := conf.ReadConf()
	err = helpers.EnsureDirectoryExists(cfg.Recorder.RecordingsDir)
	if err != nil {
		r.logger.Errorf("%v", err)
		return err
	}

	r.wg.Add(1)
	go r.record()
	return nil
}

func (r *RTSP_H264Recorder) Stop() error {
	close(r.stopCh)  // Signal the recording goroutine to stop
	r.wg.Wait()      // Wait for the recording goroutine to finish
	r.client.Close() // Close the RTSP connection
	return nil
}

func (r *RTSP_H264Recorder) sendFrame(frame image.Image) error {
	select {
	case r.eChans.FrameOut <- frame:
		return nil
	case <-r.stopCh:
		r.logger.Info("received stop signal")
		return nil
	default:
		r.logger.Info("buffer is full")
		return nil
	}
}

func (r *RTSP_H264Recorder) record() error {
	r.wg.Done()

	u, err := base.ParseURL(r.rtspURL)

	if err != nil {
		r.logger.Error("failed to parse url: %w", err)
		return err
	}
	r.logger.Info("recording...")

	// find published medias
	desc, _, err := r.client.Describe(u)
	if err != nil {
		return err
	}

	// find the H264 media and format
	var forma *format.H264
	medi := desc.FindFormat(&forma)
	if medi == nil {
		r.logger.Error("media not found")
		return nil
	}

	// setup RTP/H264 -> H264 decoder
	rtpDec, err := forma.CreateDecoder()
	if err != nil {
		r.logger.Errorf("%v", err)
		return err
	}

	// setup H264 -> MPEG-TS muxer
	mpegtsMuxer, err := newMPEGTSMuxer(forma.SPS, forma.PPS)
	if err != nil {
		return err
	}

	// setup H264 -> jpeg muxer
	frameDec, err := newH264Decoder()
	if err != nil {
		r.logger.Errorf("%v", err)
		return err
	}

	// if SPS and PPS are present into the SDP, send them to the decoder
	if forma.SPS != nil {
		frameDec.decode(forma.SPS)
	}
	if forma.PPS != nil {
		frameDec.decode(forma.PPS)
	}

	defer frameDec.close()

	// setup a single media
	_, err = r.client.Setup(desc.BaseURL, medi, 0, 0)
	if err != nil {
		return err
	}

	// called when a RTP packet arrives
	r.client.OnPacketRTP(medi, forma, func(pkt *rtp.Packet) {
		// decode timestamp
		pts, ok := r.client.PacketPTS(medi, pkt)
		if !ok {
			return
		}

		// extract access unit from RTP packets
		au, err := rtpDec.Decode(pkt)
		if err != nil {
			if err != rtph264.ErrNonStartingPacketAndNoPrevious && err != rtph264.ErrMorePacketsNeeded {
				r.logger.Errorf("%v", err)
			}
			return
		}

		// wait for an I-frame
		iFrameReceived := false
		if !iFrameReceived {
			if !h264.IDRPresent(au) {
				iFrameReceived = true
			}
		}

		// Loop over the NALUs and decode to frames.
		for _, nalu := range au {
			if err != nil {
				r.logger.Errorf("Failed to decode NALU: %v", err)
				continue // Skip this NALU if there's an error.
			}

			img, err := frameDec.decode(nalu) // Decode NALU to an image.

			// wait for a frame
			if img == nil {
				continue
			}

			// At this point, 'img' is an image.Image containing your frame.
			// You can now process, save, or stream this image as needed.
			// For example, to save as a JPEG:
			if iFrameReceived {
				err = r.sendFrame(img)
			}
			if err != nil {
				r.logger.Errorf("Failed to send frame: %v", err)
				// Handle error as needed.
			}
		}

		// encode the access unit into MPEG-TS
		err = mpegtsMuxer.encode(au, pts, r.eChans.RecordOut)
		if err != nil {
			r.logger.Errorf("%v", err)
			return
		}

	})

	// start playing
	_, err = r.client.Play(nil)
	if err != nil {
		return err
	}

	// Use a channel to receive an error from the client
	clientErrCh := make(chan error, 1)

	// Run the client's Wait in a goroutine
	go func() {
		clientErrCh <- r.client.Wait()
	}()

	select {
	case <-r.stopCh:
		r.logger.Info("received stop signal")
		return nil
	case err := <-clientErrCh:
		if err != nil {
			return err
		}
	}

	return nil
}
