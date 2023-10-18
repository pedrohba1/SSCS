package recorder

import (
	"bufio"
	"os"
	BaseLogger "sscs/logger"
	"time"

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
}

// newMPEGTSMuxer allocates a mpegtsMuxer.
func newMPEGTSMuxer(sps []byte, pps []byte) (*mpegtsMuxer, error) {
	f, err := os.Create(createChunkFileName())
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
		logger:         BaseLogger.BaseLogger.WithField("package", "recorder"),
	}, nil
}

// close closes all the mpegtsMuxer resources.
func (e *mpegtsMuxer) close() {
	e.b.Flush()
	e.f.Close()
}

func createChunkFileName() string {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	return "./recordings/feed_" + timestamp + ".ts"
}

// encode encodes a H264 access unit into MPEG-TS.
func (e *mpegtsMuxer) encode(au [][]byte, pts time.Duration) error {

	var err error
	var shouldSplit bool = false

	// Check if this Access Unit contains a keyframe and it's time to split
	for _, nalu := range au {
		typ := h264.NALUType(nalu[0] & 0x1F)
		if typ == h264.NALUTypeIDR && time.Since(e.startTimestamp) > e.chunkDuration {
			shouldSplit = true
			break
		}
	}

	if shouldSplit {
		// Close the current resources
		e.logger.Info("saving content: " + e.f.Name())
		e.b.Flush()
		e.f.Close()

		// Start a new file
		e.f, err = os.Create(createChunkFileName())
		if err != nil {
			return err
		}
		e.b = bufio.NewWriter(e.f)
		e.w = mpegts.NewWriter(e.b, []*mpegts.Track{e.track})
		e.startTimestamp = time.Now()
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
			e.sps = nalu
			continue

		case h264.NALUTypePPS:
			e.pps = nalu
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
		au = append([][]byte{e.sps, e.pps}, au...)
	}

	var dts time.Duration

	if e.dtsExtractor == nil {
		// skip samples silently until we find one with a IDR
		if !idrPresent {
			return nil
		}
		e.dtsExtractor = h264.NewDTSExtractor()
	}

	dts, err = e.dtsExtractor.Extract(au, pts)
	if err != nil {
		return err
	}

	// encode into MPEG-TS
	return e.w.WriteH26x(e.track, durationGoToMPEGTS(pts), durationGoToMPEGTS(dts), idrPresent, au)
}
