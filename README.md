### Go Media Framework.
FFmpeg libav* Go bindings.

#### Status `unfinished`.
It covers very basic avformat, avcodec and swscale features.    
More bindings and cool features are coming soon.

#### Install
You have to build ffmpeg with option `--enable-shared` or install shared libraries with your package manager (if it has them).

Ensure that PKG_CONFIG_PATH contains path to ffmpeg's pkgconfig folder.

```sh
# check it by running
pkg-config --libs libavformat
```

It should print valid path to the avformat library.  

Now, just run

```sh
go get github.com/3d0c/gmf
```

#### Usage
Please see examples and tests. 

