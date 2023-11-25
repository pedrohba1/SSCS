#!/bin/sh

# Wait for mediamtx to start up. This is a very simple check.
/mediamtx &

echo "Waiting for mediamtx to initialize..."

sleep 2

echo "initializing ffpmeg stream..."

# Start the ffmpeg stream
# ffmpeg -re -stream_loop -1 -i /samples/sp1.mp4 -c copy -f rtsp rtsp://localhost:8554/mystream

ffmpeg -re -stream_loop -1 -i ./samples/sp1_no_bf.mp4 \
 -vcodec copy -c:a libopus  \
 -f rtsp rtsp://localhost:8554/mystream
