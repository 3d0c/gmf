### Examples
##### encoding.go
is a port of ffmpeg/doc/examples/encoding-decoding.c

```sh
go run encoding.go
```

##### encoding-multiple.go
is an experimental stuff, which is producing three different output at once. It creates three workers with different codecs settings and passes them synthetic generated frames. As a result you will get three files, encoded with mpeg1, mpeg2 and mpeg4.

```sh
go run encoding-multiple.go 
```

##### transcode.go 
is a simple transcoder. It gets two best streams (video and audio) from input and converts them to mpeg4 and aac.

```sh
go run transcode [input] [output.mp4]
```

#### video-to-jpeg.go
extracts frames and encodes them to jpeg files. 

```sh
# if no args specified, default test-sample.mp4 will be used
go run video-to-jpeg.go [input_video]
```

#### video-to-jpeg-p.go
a bit faster parallel version.
  
