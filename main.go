package main

import (
	"bytes"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/disintegration/gift"
	"github.com/otiai10/gosseract/v2"
)

func main() {
	if len(os.Args) < 2 {
		panic("missing image path")
	}
	if len(os.Args) < 3 {
		panic("missing blur level number")
	}
	if len(os.Args) < 4 {
		panic("missing text confidence number")
	}

	path := os.Args[1]
	blurLevel := os.Args[2]
	minConfidence := os.Args[3]

	err := blur(path, blurLevel, minConfidence)
	if err != nil {
		panic(err)
	}
}

func blur(path string, blurLevelStr, minConfidenceStr string) error {
	blurLevel, err := strconv.Atoi(blurLevelStr)
	if err != nil {
		return err
	}

	minConfidence, err := strconv.Atoi(minConfidenceStr)
	if err != nil {
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	img, err := png.Decode(bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	rgba := image.NewRGBA(img.Bounds())
	gift.New().Draw(rgba, img)

	client := gosseract.NewClient()
	defer client.Close()
	client.SetImageFromBytes(data)
	boxes, err := client.GetBoundingBoxes(gosseract.RIL_WORD)
	if err != nil {
		return err
	}

	for _, box := range boxes {
		if box.Confidence < float64(minConfidence) {
			continue
		}
		bounds := box.Box.Bounds()
		gift.New(gift.GaussianBlur(float32(blurLevel))).DrawAt(
			rgba,
			rgba.SubImage(box.Box),
			image.Pt(bounds.Min.X, bounds.Min.Y),
			gift.OverOperator,
		)
	}

	newPath := strings.TrimSuffix(path, filepath.Ext(path)) + "_blurred.png"
	f, err := os.Create(newPath)
	if err != nil {
		return err
	}
	err = png.Encode(f, rgba)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}
