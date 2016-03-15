### Go FFmpeg Bindings 

#### Status: `beta`
It covers very basic avformat, avcodec and swscale features.    
More bindings and cool features are coming soon.

#### Installation
##### Prerequisites
Current master branch requires `go 1.6`.  
Older versions available in branches go1.2 and go1.5.

##### Build
build lastest version of ffmpeg, obtained from [https://github.com/FFmpeg/FFmpeg](https://github.com/FFmpeg/FFmpeg)  
There is one required option, which is disabled by default, you should turn on: `--enable-shared`  

E.g.:

```sh
./configure --prefix=/usr/local/ffmpeg --enable-shared
make
make install
```

Add pkgconfig path:

```sh
export PKG_CONFIG_PATH=$PKG_CONFIG_PATH:/usr/local/ffmpeg/lib/pkgconfig/
```

Ensure, that `PKG_CONFIG_PATH` contains path to ffmpeg's pkgconfig folder.

```sh
# check it by running
pkg-config --libs libavformat
```

It should print valid path to the avformat library.  

Now, just run

```sh
go get github.com/3d0c/gmf
```

##### Other methods
This package uses pkg-config way to obtain flags, includes and libraries path, so if you have ffmpeg installed, just ensure, that your installation has them (pkgconfig/ folder with proper `pc` files).

#### Usage
Please see [examples](examples/) and tests. 

#### Support and Contribution
If something doesn't work, just fix it. Do not hesitate to pull request.

#### Credits
I borrowed the name from project, abandoned on code.google.com/p/gmf. Original code is available here in intitial commit from 03 Apr 2013.
