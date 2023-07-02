package main

import (
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/draw"
	"image/gif"
	"io/ioutil"
	"os"
)

const (
	CHARS        = "‎ "
	ESC          = "\u001b"
	CSI          = "["
	SGR_END      = ""
	CHA_END      = "G"
	CUP_END      = "H"
	ED_END       = "J"
	RESET_CURSOR = ESC + CSI + "1;1" + CUP_END
	HIDE_CURSOR  = ESC + CSI + "?25l"
	RESET_DISPLAY = ESC + CSI + "0" + SGR_END + ESC + CSI + "2" + ED_END
)

func getChar(brightness float64) string {
	index := int(brightness * (float64(len(CHARS)) - 1))
	return string(CHARS[index])
}

func getRGBEscape(r, g, b uint8) string {
	return fmt.Sprintf("\u001b[48;2;%d;%d;%dm\u001b[38;2;%d;%d;%dm", r, g, b, r, g, b)
}
func imageToText(image image.Image) string {
	text := ""
	bounds := image.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := image.At(x, y).RGBA()
			r >>= 8
			g >>= 8
			b >>= 8
			brightness := (0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)) / 255.0
			char := getChar(brightness)
			fgEscape := getRGBEscape(uint8(r), uint8(g), uint8(b))
			text += fgEscape + char + SGR_END
		}
		text += "\n"
		text += "<<sleep(3)>>"
	}
	text = text[:len(text)-14]
	return text
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Usage: go run main.go <gif_path>")
		return
	}

	gifPath := args[0]
	file, err := os.Open(gifPath)
	if err != nil {
		fmt.Printf("Error opening GIF file: %v\n", err)
		return
	}
	defer file.Close()

	g, err := gif.DecodeAll(file)
	if err != nil {
		fmt.Printf("Error decoding GIF: %v\n", err)
		return
	}

	frames := g.Image
	numFrames := len(frames)
	resizedFrames := make([]*image.RGBA, numFrames)
	for i, frame := range frames {
		resizedFrames[i] = resizeImage(frame, 80, 24)
	}

	text := HIDE_CURSOR

	for _, frame := range resizedFrames {
		imageText := imageToText(frame)
		text += RESET_CURSOR + imageText
	}

	text += RESET_DISPLAY

	err = ioutil.WriteFile("evo.txt", []byte(text), 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}
	fmt.Println("Output written to evo.txt")
}

func resizeImage(img image.Image, newWidth, newHeight int) *image.RGBA {
	resized := resize.Resize(uint(newWidth), uint(newHeight), img, resize.Lanczos3)
	bounds := resized.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, resized, image.Point{0, 0}, draw.Src)
	return rgba
}
