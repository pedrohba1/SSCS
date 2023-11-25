# Start from the mediamtx image with ffmpeg included
FROM bluenviron/mediamtx:1.2.0-ffmpeg

# Copy your media files into the container
COPY ./samples /samples

# Copy the start script into the container
COPY ./start.sh /start.sh
RUN chmod +x /start.sh

# # Use the start script as the entry point
ENTRYPOINT ["/start.sh"]
