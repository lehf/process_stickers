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
		files, filesErr := ioutil.ReadDir(filePath)
		if filesErr != nil {
			log.Println(filePath)
		}
		for _, file := range files {
			processImage(filePath + "/" + file.Name())
		}

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
					w, h := img.Bounds().Dx(), img.Bounds().Dy()
					if w == 0 || h == 0 {
						log.Println("Invalid image size:", w, h)
					} else {
						if w == h {
							output, _ := os.Create(filePath + "_resized.png")
							defer output.Close()
							dst := image.NewRGBA(image.Rect(0, 0, 120, 120))
							draw.CatmullRom.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)
							log.Println("Image resizing:", output)
							png.Encode(output, dst)
						}
						log.Println(filePath+"Image size:", w, h)
					}
				} else if format == "gif" {
					pic.Seek(0, 0)
					output, _ := os.Create(filePath + "_resized.gif")
					defer output.Close()
					tx := func(m image.Image) image.Image {
						return resize.Resize(240, 240, m, resize.Lanczos3)
					}

					err = ProcessGif(output, pic, tx)
					if err != nil {
						log.Println("error processing gif: ", err)
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
	myWindow := myApp.NewWindow("Process Sticker")
	content := widget.NewButton("Click to Select Folder", func() {
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

			children, err := list.List()
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			folderPath := list.Path()
			log.Println(folderPath)
			_, filesErr := os.Stat(folderPath)
			if filesErr != nil {
				dialog.ShowError(filesErr, myWindow)
				return
			}
			processImage(folderPath)
			out := fmt.Sprintf("Folder %s (%d children):\n%s", list.Name(), len(children), list.String())
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
