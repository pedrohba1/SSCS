package recorder

import (
	"os"
	"sync"

	BaseLogger "sscs/logger"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/bluenviron/gortsplib/v4/pkg/format/rtph264"
	"github.com/bluenviron/gortsplib/v4/pkg/url"
	"github.com/pion/rtp"
	"github.com/sirupsen/logrus"
)

type RTSPRecorder struct {
	rtspURL string
	client  *gortsplib.Client
	logger  *logrus.Entry

	wg       sync.WaitGroup
	recordIn chan<- RecordedEvent
	stopCh   chan struct{}
}

func ensureDirectoryExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755) // 0755 means everyone can read, owner can write
	}
	return nil
}

func NewRTSPRecorder(rtspURL string, recordChan chan RecordedEvent) *RTSPRecorder {
	r := &RTSPRecorder{
		rtspURL:  rtspURL,
		recordIn: recordChan,
		stopCh:   make(chan struct{}),
	}
	r.setupLogger()

	return r
}

func (r *RTSPRecorder) setupLogger() {
	r.logger = BaseLogger.BaseLogger.WithField("package", "recorder")
}

func (r *RTSPRecorder) Start() error {
	u, err := url.Parse(r.rtspURL)

	r.client = &gortsplib.Client{}

	// connect to the server
	err = r.client.Start(u.Scheme, u.Host)
	if err != nil {
		r.logger.Error("failed to start RTSP client: %w", err)
		return err
	}

	r.wg.Add(1)
	go r.record()
	return nil
}

func (r *RTSPRecorder) Stop() error {
	close(r.stopCh)  // Signal the recording goroutine to stop
	r.wg.Wait()      // Wait for the recording goroutine to finish
	r.client.Close() // Close the RTSP connection
	return nil
}

func (r *RTSPRecorder) record() error {
	defer r.wg.Done()

	u, err := url.Parse(r.rtspURL)

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
	var form *format.H264
	medi := desc.FindFormat(&form)
	if medi == nil {
		r.logger.Warn("media not found")
		return nil
	}

	// setup RTP/H264 -> H264 decoder
	rtpDec, err := form.CreateDecoder()
	if err != nil {
		r.logger.Errorf("%v", err)
		return err
	}

	// Ensure the recordings directory exists
	err = ensureDirectoryExists("./recordings")
	if err != nil {
		r.logger.Errorf("%v", err)
		return err
	}

	// setup H264 -> MPEG-TS muxer
	mpegtsMuxer, err := newMPEGTSMuxer(form.SPS, form.PPS)
	if err != nil {
		return err
	}

	// setup a single media
	_, err = r.client.Setup(desc.BaseURL, medi, 0, 0)
	if err != nil {
		return err
	}

	// called when a RTP packet arrives
	r.client.OnPacketRTP(medi, form, func(pkt *rtp.Packet) {
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

		// encode the access unit into MPEG-TS
		err = mpegtsMuxer.encode(au, pts, r.recordIn)
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
