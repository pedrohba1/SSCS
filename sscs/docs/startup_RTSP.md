# Startup of RTSP server

The idea here is that SCSS simply processes and stores the image from a
RTSP server. Let's start by initiating an RTSP server:

```
# run in root of this repo
docker run --rm -d -it --network=host -v $PWD/scss/mediamtx.yml bluenviron/mediamtx
```

The script above creates a docker container with [bluenviron/mediamtx image](https://hub.docker.com/r/bluenviron/mediamtx) which is a server and proxy that allows users to publish, read and proxy live video and audio streams. Basically,

And then it's possible to push a stream to the server with:

```
# run in root of this repo
ffmpeg -re -stream_loop -1 -i ./sscs/samples/sp1.mp4 -c copy -f rtsp rtsp://localhost:8554/mystream
```

To read the stream, for checking if it works, you can run the command below. (ffplay comes with ffmpeg)

```
ffplay rtsp://localhost:8554/mystream
```

You can start multiple RTSP feeds using this approach. Or either use a single RTSP server to create multiple streams,
just by chaning the `mystream` path
