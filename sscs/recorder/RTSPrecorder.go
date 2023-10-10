package recorder

import (
	"log"
	"sync"

	"github.com/aler9/gortsplib/pkg/rtpcodecs/rtph264"
	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/bluenviron/gortsplib/v4/pkg/url"
	"github.com/pion/rtp"
)

type RTSPRecorder struct {
	rtspURL string
	client  *gortsplib.Client
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

func NewRTSPRecorder(rtspURL string) *RTSPRecorder {
	return &RTSPRecorder{
		rtspURL: rtspURL,
		stopCh:  make(chan struct{}),
	}
}

func (r *RTSPRecorder) Start() error {
	// parse URL
	u, err := url.Parse(r.rtspURL)
	if err != nil {
		panic(err)
	}

	r.client = &gortsplib.Client{}

	// connect to the server
	err = r.client.Start(u.Scheme, u.Host)
	if err != nil {
		panic(err)
	}
	defer r.client.Close()

	r.record(u)
	return nil
}

func (r *RTSPRecorder) Stop() error {
	close(r.stopCh)  // Signal the recording goroutine to stop
	r.wg.Wait()      // Wait for the recording goroutine to finish
	r.client.Close() // Close the RTSP connection
	return nil
}

// continuar à partir desse exemplo:

func (r *RTSPRecorder) record(u *url.URL) {
	defer r.wg.Done()

	// find published medias
	desc, _, err := r.client.Describe(u)
	if err != nil {
		panic(err)
	}

	// find the H264 media and format
	var forma *format.H264
	medi := desc.FindFormat(&forma)
	if medi == nil {
		panic("media not found")
	}

	// setup RTP/H264 -> H264 decoder
	rtpDec, err := forma.CreateDecoder()
	if err != nil {
		panic(err)
	}

	// setup H264 -> MPEG-TS muxer
	mpegtsMuxer, err := newMPEGTSMuxer(forma.SPS, forma.PPS)
	if err != nil {
		panic(err)
	}

	// setup a single media
	_, err = r.client.Setup(desc.BaseURL, medi, 0, 0)
	if err != nil {
		panic(err)
	}

	// called when a RTP packet arrives
	r.client.OnPacketRTP(medi, forma, func(pkt *rtp.Packet) {
		// decode timestamp
		pts, ok := r.client.PacketPTS(medi, pkt)
		if !ok {
			log.Printf("waiting for timestamp")
			return
		}

		// extract access unit from RTP packets
		au, err := rtpDec.Decode(pkt)
		if err != nil {
			if err != rtph264.ErrNonStartingPacketAndNoPrevious && err != rtph264.ErrMorePacketsNeeded {
				log.Printf("ERR: %v", err)
			}
			return
		}

		// encode the access unit into MPEG-TS
		err = mpegtsMuxer.encode(au, pts)
		if err != nil {
			log.Printf("ERR: %v", err)
			return
		}

		log.Printf("saved TS packet")
	})

	// start playing
	_, err = r.client.Play(nil)
	if err != nil {
		panic(err)
	}

	// wait until a fatal error
	panic(r.client.Wait())

}
