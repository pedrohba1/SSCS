package recorder

import (
	"bufio"
	"os"
	"time"

	"github.com/pedrohba1/SSCS/services/conf"
	BaseLogger "github.com/pedrohba1/SSCS/services/logger"

	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/pkg/formats/mpegts"
	"github.com/sirupsen/logrus"
)

func durationGoToMPEGTS(v time.Duration) int64 {
	return int64(v.Seconds() * 90000)
}

// mpegtsMuxer allows to save a H264 stream into a MPEG-TS file.
type mpegtsMuxer struct {
	sps []byte
	pps []byte

	f              *os.File
	b              *bufio.Writer
	w              *mpegts.Writer
	startTimestamp time.Time
	chunkDuration  time.Duration
	track          *mpegts.Track
	dtsExtractor   *h264.DTSExtractor
	logger         *logrus.Entry
	recordingsDir string

	recordOut chan RecordedEvent
}

// newMPEGTSMuxer allocates a mpegtsMuxer.
func newMPEGTSMuxer(sps []byte, pps []byte) (*mpegtsMuxer, error) {
	 
	cfg, _ := conf.ReadConf()
	f, err := os.Create(createChunkFileName(cfg.Recorder.RecordingsDir))
	if err != nil {
		return nil, err
	}
	b := bufio.NewWriter(f)
	track := &mpegts.Track{
		Codec: &mpegts.CodecH264{},
	}
	w := mpegts.NewWriter(b, []*mpegts.Track{track})

	return &mpegtsMuxer{
		sps:            sps,
		pps:            pps,
		f:              f,
		b:              b,
		w:              w,
		startTimestamp: time.Now(),
		chunkDuration:  8 * time.Second,
		track:          track,
		recordingsDir: cfg.Recorder.RecordingsDir,
		logger:         BaseLogger.BaseLogger.WithField("package", "recorder"),
	}, nil
}

// close closes all the mpegtsMuxer resources.
func (e *mpegtsMuxer) close() {
	e.b.Flush()
	e.f.Close()
}

func createChunkFileName(recordingsDir string) string {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	return recordingsDir + "/feed_" + timestamp + ".ts"
}

// encode encodes a H264 access unit into MPEG-TS.
func (mux *mpegtsMuxer) encode(au [][]byte, pts time.Duration, recordIn chan<- RecordedEvent) error {

	var err error
	var shouldSplit bool = false

	// Check if this Access Unit contains a keyframe and it's time to split
	for _, nalu := range au {
		typ := h264.NALUType(nalu[0] & 0x1F)
		if typ == h264.NALUTypeIDR && time.Since(
			mux.startTimestamp) >
			mux.chunkDuration {
			shouldSplit = true
			break
		}
	}

	if shouldSplit {
		// Close the current resources
		mux.logger.Info("saving content: " + mux.f.Name())
		recordIn <- RecordedEvent{Path: mux.f.Name(), EndTime: time.Now()}

		mux.b.Flush()
		mux.f.Close()

		// Start a new file
		mux.f, err = os.Create(createChunkFileName(mux.recordingsDir))
		if err != nil {
			return err
		}
		mux.b = bufio.NewWriter(mux.f)
		mux.w = mpegts.NewWriter(mux.b, []*mpegts.Track{mux.track})
		mux.startTimestamp = time.Now()
	}
	// prepend an AUD. This is required by some players
	filteredAU := [][]byte{
		{byte(h264.NALUTypeAccessUnitDelimiter), 240},
	}

	nonIDRPresent := false
	idrPresent := false

	for _, nalu := range au {
		typ := h264.NALUType(nalu[0] & 0x1F)
		switch typ {
		case h264.NALUTypeSPS:
			mux.sps = nalu
			continue

		case h264.NALUTypePPS:
			mux.pps = nalu
			continue

		case h264.NALUTypeAccessUnitDelimiter:
			continue

		case h264.NALUTypeIDR:
			idrPresent = true

		case h264.NALUTypeNonIDR:
			nonIDRPresent = true
		}

		filteredAU = append(filteredAU, nalu)
	}

	au = filteredAU

	if len(au) <= 1 || (!nonIDRPresent && !idrPresent) {
		return nil
	}

	// add SPS and PPS before every access unit that contains an IDR
	if idrPresent {
		au = append([][]byte{mux.sps, mux.pps}, au...)
	}

	var dts time.Duration

	if mux.dtsExtractor == nil {
		// skip samples silently until we find one with a IDR
		if !idrPresent {
			return nil
		}
		mux.dtsExtractor = h264.NewDTSExtractor()
	}

	dts, err = mux.dtsExtractor.Extract(au, pts)
	if err != nil {
		return err
	}

	// encode into MPEG-TS
	return mux.w.WriteH26x(mux.track, durationGoToMPEGTS(pts), durationGoToMPEGTS(dts), idrPresent, au)
}
