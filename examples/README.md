## Examples

NOTE. Examples not included in this list are old and might not working. (Under developing now)

### video-to-image.go

Save video frames as images. Generates sortable images in corresponding format inside `./tmp` directory (useful for `images-to-video.go` example).

Options:  
  
  ```
  -ext string
    	destination type, e.g.: png, jpg, tiff, whatever encoder you have (default
    	 "png")
  -src string
    	source video (default "tests-sample.mp4")
  ```    

### video-to-goImage.go

Same as above but using Golang `image/jpeg` package. 

Options:  
  
  ```
  -src string
    	source video (default "tests-sample.mp4")
  ```    
  
### images-to-video.go

Generate video from sequence of images.

Options:  
  
  ```
  -dst string
    	destination file (default "result.mp4")
  -src string
    	source images folder (default "./tmp")
  ```
  
### watermark.go

Add watermark to video using complex filter.

Options:

```
  -dst string
    	destination file, e.g. -dst=result.mp4
  -src value
    	source files, e.g.: -src=1.mp4 -src=image.png
```
_Note, applying image should be the last in "-src" option._

### mp4s-to-flv.go

Concat video files into single "flv".

```
  -dst string
    	destination file, e.g. -dst=result.mp4
  -src value
    	source files, e.g.: -src=1.mp4 -src=2.mp4
```


