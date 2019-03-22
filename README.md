### Go FFmpeg Bindings 

#### Installation
##### Prerequisites
Current master branch supports all major Go versions, starting from 1.6.   

##### Build/install FFmpeg
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

#### Docker containers
Thanks to [@ergoz](https://github.com/ergoz) you can try a docker container [riftbit/ffalpine](https://hub.docker.com/r/riftbit/ffalpine)

Thanks to [@denismakogon](https://github.com/denismakogon) there is one more project, worth to mention
[https://github.com/denismakogon/ffmpeg-debian](https://github.com/denismakogon/ffmpeg-debian)

#### Usage
Please see [examples](examples/).

#### Support and Contribution
If something doesn't work, just fix it. Do not hesitate to pull request.

#### Credits
I borrowed the name from project, abandoned on code.google.com/p/gmf. Original code is available here in intitial commit from 03 Apr 2013.
