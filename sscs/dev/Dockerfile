FROM bluenviron/mediamtx

# Install ffmpeg if it's not already in the image (optional, only if the base image doesn't have ffmpeg)
RUN apt-get update && apt-get install -y bash ffmpeg

# Copy necessary files
COPY  ./samples/sp1.mp4 /sscs/samples/sp1.mp4
COPY ./mediamtx.yml /sscs/mediamtx.yml


# Set the entry point to ffmpeg command
# CMD ["ffmpeg", "-re", "-stream_loop", "-1", "-i", "./samples/sp1.mp4", "-c", "copy", "-f", "rtsp", "rtsp://localhost:8554/mystream"]
