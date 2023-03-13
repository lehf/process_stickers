package main

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/nfnt/resize"
	"golang.org/x/image/draw"
)

func processImage(filePath string) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Println(err, "filePath")
	}
	if fileInfo.IsDir() {
		// files, filesErr := ioutil.ReadDir(filePath)
		// if filesErr != nil {
		// 	log.Println(filePath)
		// }
		// for _, file := range files {
		// 	processImage(filePath + "/" + file.Name())
		// }

	} else {
		pic, picerr := os.Open(filePath)

		if picerr != nil {
			log.Println("Cannot read file:", picerr)
		} else {
			img, format, imgerr := image.Decode(pic)
			if imgerr != nil {
				log.Println("Cannot decode file:", filePath)
			} else {
				if format == "png" || format == "jpg" || format == "jpeg" {
					log.Println("Image size:", img.Bounds().Dx())
					// w, h := img.Bounds().Dx(), img.Bounds().Dy()
					// if w == 0 || h == 0 {
					// 	log.Println("Invalid image size:", w, h)
					// } else {
					// 	if w == h {
					// 		output, _ := os.Create(filePath + "_resized.png")
					// 		defer output.Close()
					// 		dst := image.NewRGBA(image.Rect(0, 0, 120, 120))
					// 		draw.CatmullRom.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)
					// 		log.Println("Image resizing:", output)
					// 		png.Encode(output, dst)
					// 	}
					// 	log.Println(filePath+"Image size:", w, h)
					// }
				} else if format == "gif" {
					pic.Seek(0, 0)
					otpth := strings.TrimSuffix(filePath, filepath.Ext(filePath))
					thuotpth := strings.TrimSuffix(filePath, filepath.Ext(filePath))
					output, _ := os.Create(otpth + "_resized.gif")
					thumboutput, _ := os.Create(thuotpth + "_thumb.png")
					defer output.Close()
					defer thumboutput.Close()

					tx := func(m image.Image) image.Image {
						return resize.Resize(240, 240, m, resize.Lanczos3)
					}

					err1 := ProcessGif(output, pic, tx)
					pic.Seek(0, 0)
					err2 := Thumbnail(thumboutput, pic)

					fi, fierr := os.Stat(otpth + "_resized.gif")
					if fierr != nil {
						log.Println("error processing gif: ", err1)
					}
					// get the size
					size := fi.Size()
					if size > 500000 {
						log.Println("Image File size:", size)
					}
					if err1 != nil {
						log.Println("error processing gif: ", err1)
					}
					if err2 != nil {
						log.Println("error thumb gif: ", err2)
					}
				} else {
					log.Println("image name:", format)
				}
			}
		}
		defer pic.Close()

	}

}
func imageToPaletted(img image.Image, p color.Palette) *image.Paletted {
	b := img.Bounds()
	pm := image.NewPaletted(b, p)
	draw.FloydSteinberg.Draw(pm, b, img, image.ZP)
	return pm
}

// TransformFunc is a function that transforms an image.
type TransformFunc func(image.Image) image.Image

func Thumbnail(w io.Writer, r io.Reader) error {
	im, err := gif.DecodeAll(r)
	if err != nil {
		return err
	}
	firstFrame := im.Image[0]
	b := image.Rect(0, 0, 120, 120)
	dst := image.NewRGBA(b)
	draw.CatmullRom.Scale(dst, dst.Rect, firstFrame, firstFrame.Bounds(), draw.Over, nil)
	log.Println("Gif thumbing:", w)
	return png.Encode(w, dst)
}

// ProcessGif the GIF read from r, applying transform to each frame, and writing
// the result to w.
func ProcessGif(w io.Writer, r io.Reader, transform TransformFunc) error {
	if transform == nil {
		_, err := io.Copy(w, r)
		return err
	}

	// Decode the original gif.
	im, err := gif.DecodeAll(r)
	if err != nil {
		return err
	}

	// Create a new RGBA image to hold the incremental frames.
	firstFrame := im.Image[0].Bounds()
	// if firstFrame.Dx() != firstFrame.Dy() {
	// 	return errors.New("first frame must be square")
	// }
	b := image.Rect(0, 0, firstFrame.Dx(), firstFrame.Dy())
	img := image.NewRGBA(b)

	// Resize each frame.
	for index, frame := range im.Image {
		bounds := frame.Bounds()
		previous := img
		draw.Draw(img, bounds, frame, bounds.Min, draw.Over)
		im.Image[index] = imageToPaletted(transform(img), frame.Palette)

		switch im.Disposal[index] {
		case gif.DisposalBackground:
			img = image.NewRGBA(b)
		case gif.DisposalPrevious:
			img = previous
		}
	}
	im.Config.Width = im.Image[0].Bounds().Max.X
	im.Config.Height = im.Image[0].Bounds().Max.Y

	return gif.EncodeAll(w, im)
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("动图处理")
	content := widget.NewButton("Select folder", func() {
		dialog.ShowFolderOpen(func(list fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			log.Println(list)
			if list == nil {
				log.Println("Cancelled")
				return
			}

			// children, err := list.List()
			// if err != nil {
			// 	dialog.ShowError(err, myWindow)
			// 	return
			// }
			folderPath := list.Path()
			log.Println(folderPath)
			_, filesErr := os.Stat(folderPath)
			if filesErr != nil {
				dialog.ShowError(filesErr, myWindow)
				return
			}
			fileInfo, errfilePath := os.Stat(folderPath)
			if errfilePath != nil {
				log.Println(errfilePath, "filePath")
			}
			if fileInfo.IsDir() {
				files, filesErr := ioutil.ReadDir(folderPath)
				if filesErr != nil {
					log.Println(folderPath)
				}
				for _, file := range files {
					processImage(folderPath + "/" + file.Name())
				}

			}
			// processImage(folderPath)
			out := fmt.Sprintf("Finished folder")
			dialog.ShowInformation("Folder Open", out, myWindow)
		}, myWindow)
	})
	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(480, 360))
	myWindow.Show()
	myApp.Run()
	tidyUp()
}

func tidyUp() {
	fmt.Println("Exited")
}
