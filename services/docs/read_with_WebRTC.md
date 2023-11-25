# Reading a server with webRTC

This configuration needs to be added in order to use the server inside a container.
Otherwise, webRTC in a development environment simply will not work.

```
# public IP of the server
webrtcICEHostNAT1To1IPs: [192.168.x.x]
# any port of choice
webrtcICEUDPMuxAddress: :8189
```


It is also necessary to remove B-frames in case of reading with WebRTC because this protocol
does not support it. In local development, you can do it like this:

```
ffmpeg -i sp1.mp4 -c:v libx264 -bf 0  sp1_no_bf.mp4
```

If there is no sound in the stream, you can transcode the audio.

```
#run in root
ffmpeg -re -stream_loop -1 -i ./sscs/dev/samples/sp1_no_bf.mp4 \
 -vcodec copy -c:a libopus  \
 -f rtsp rtsp://localhost:8554/mystream
```