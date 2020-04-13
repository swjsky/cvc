package handler

import (
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"strconv"

	"github.com/nfnt/resize"

	"loreal.com/dit/module/modules/resource"
)

//ImageResizeHandler - mime handler to resize image
type ImageResizeHandler struct {
	Mimes []string
}

//NewImageResizeHandler - create a new mime handler to resize image
func NewImageResizeHandler() resource.MimeHandler {
	return resource.MimeHandler(ImageResizeHandler{
		[]string{
			"image/jpeg",
			"image/jpg",
			"image/png",
		},
	})
}

//Match - whether mime can be processed by this handler
func (h ImageResizeHandler) Match(mime string) bool {
	for _, m := range h.Mimes {
		if m == mime {
			return true
		}
	}
	return false
}

//Process - process resource by this handler
func (h ImageResizeHandler) Process(sourceFile, targetFile, mime, options string) error {
	var width uint
	if options == "thumb" {
		width = 480
	} else {
		if w, err := strconv.Atoi(options); err == nil {
			width = uint(w)
		} else {
			return errors.New("Invalid width")
		}
	}
	fin, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer fin.Close()
	origin, _, err := image.Decode(fin)
	var canvas image.Image
	if options == "thumb" {
		canvas = resize.Thumbnail(width, width, origin, resize.Lanczos3)
	} else {
		canvas = resize.Resize(width, 0, origin, resize.Lanczos3)
	}
	fout, foutErr := os.Create(targetFile)
	if foutErr != nil {
		return foutErr
	}
	defer fout.Close()
	switch mime {
	case "image/jpeg", "image/jpg":
		jpeg.Encode(fout, canvas, &jpeg.Options{Quality: 80})
		break
	case "image/png":
		png.Encode(fout, canvas)
	}
	return nil
}
