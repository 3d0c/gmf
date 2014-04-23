### Go Media Framework 
It is a FFmpeg libav* Go bindings. Just a wrapper.  

#### Status: `beta`
It covers very basic avformat, avcodec and swscale features.    
More bindings and cool features are coming soon.

#### Install
##### Recommended way:  
build the lastest version of ffmpeg, obtained from [https://github.com/FFmpeg/FFmpeg](https://github.com/FFmpeg/FFmpeg)  
There is one required option, which is disabled by default, you should turn on: `--enable-shared`  
E.g.

```sh
./configure --prefix=/usr/local/ffmpeg --enable-shared
make
make install
```

Ensure that PKG_CONFIG_PATH contains path to ffmpeg's pkgconfig folder.

```sh
# check it by running
pkg-config --libs libavformat
```

It should print valid path to the avformat library.  

Add it, if needed:

```sh
export PKG_CONFIG_PATH=$PKG_CONFIG_PATH:/usr/local/ffmpeg/lib/pkgconfig/
```

Now, just run

```sh
go get github.com/3d0c/gmf
```

#### Usage
Please see [examples](examples/) and tests. 

#### Support and Contribution
If something doesn't work, just fix it. Do not hesitate to pull request.

#### Credits
I borrowed the name from project, abandoned on code.google.com/p/gmf. Original code is available here in intitial commit from 03 Apr 2013.