package recorder

import (
	"log"
	"os"
	"sync"

	"github.com/aler9/gortsplib"
)

type RTSPRecorder struct {
	rtspURL    string
	outputFile string
	conn       *gortsplib.Client
	file       *os.File
	stopCh     chan struct{}
	wg         sync.WaitGroup
}

func NewRTSPRecorder(rtspURL, outputFile string) *RTSPRecorder {
	return &RTSPRecorder{
		rtspURL:    rtspURL,
		outputFile: outputFile,
		stopCh:     make(chan struct{}),
	}
}

func (r *RTSPRecorder) Start() error {
	// Connect to the RTSP server
	conn, err := gortsplib.DialRead(r.rtspURL)
	if err != nil {
		return err
	}
	r.conn = conn

	// Open output file
	file, err := os.Create(r.outputFile)
	if err != nil {
		return err
	}
	r.file = file

	// Start recording in a goroutine
	r.wg.Add(1)
	go r.record()

	return nil
}

func (r *RTSPRecorder) Stop() error {
	close(r.stopCh) // Signal the recording goroutine to stop
	r.wg.Wait()     // Wait for the recording goroutine to finish
	r.conn.Close()  // Close the RTSP connection
	r.file.Close()  // Close the output file
	return nil
}

// continuar à partir desse exemplo:

//https://github.com/bluenviron/gortsplib/blob/main/examples/client-read-format-h264-save-to-disk/main.go

func (r *RTSPRecorder) record() {
	defer r.wg.Done()

	for {
		select {
		case <-r.stopCh:
			return

		default:
			frame, err := r.conn.ReadFrame()
			if err != nil {
				log.Println("Error reading frame:", err)
				return
			}

			if frame.IsH264() {
				r.file.Write(frame.Payload)
			}
		}
	}
}
