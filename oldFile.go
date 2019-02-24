package main

import (
  "fmt"
  "os"
  "log"
  "path/filepath"
  "io/ioutil"
  "time"
  "net/http"
  "strings"
  "github.com/discordapp/lilliput"
)

var EncodeOptions = map[string]map[int]int{
	".jpeg": map[int]int{lilliput.JpegQuality: 85},
	".png":  map[int]int{lilliput.PngCompression: 7},
	".webp": map[int]int{lilliput.WebpQuality: 85},
}

func main() {
	var outputWidth int = 250
	var outputHeight int = 250
	var stretch bool = false
	var url string = "https://upload.wikimedia.org/wikipedia/commons/4/41/Sunflower_from_Silesia2.jpg"
	var outputFilename string = filepath.Base(url)

	resp, err := http.Get(url)
	if err != nil {
	  log.Fatalf("http.Get => %v", err.Error())
	}

	inputBuf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("failed to read input file, %s\n", err)
		os.Exit(1)
	}

	decoder, err := lilliput.NewDecoder(inputBuf)
	if err != nil {
		fmt.Printf("error decoding image, %s\n", err)
		os.Exit(1)
	}
	defer decoder.Close()


	header, err := decoder.Header()
	if err != nil {
		fmt.Printf("error reading image header, %s\n", err)
		os.Exit(1)
	}

	if decoder.Duration() != 0 {
		fmt.Printf("duration: %.2f s\n", float64(decoder.Duration())/float64(time.Second))
	}
	ops := lilliput.NewImageOps(8192)
	defer ops.Close()


	outputImg := make([]byte, 250*1024*1024)

	outputType := "." + strings.ToLower(decoder.Description())

	if outputWidth == 0 {
		outputWidth = header.Width()
	}

	if outputHeight == 0 {
		outputHeight = header.Height()
	}

	resizeMethod := lilliput.ImageOpsFit
	if stretch {
		resizeMethod = lilliput.ImageOpsResize
	}

	if outputWidth == header.Width() && outputHeight == header.Height() {
		resizeMethod = lilliput.ImageOpsNoResize
	}
	opts := &lilliput.ImageOptions{
		FileType:             outputType,
		Width:                outputWidth,
		Height:               outputHeight,
		ResizeMethod:         resizeMethod,
		NormalizeOrientation: true,
		EncodeOptions:        EncodeOptions[outputType],
	}

	outputImg, err = ops.Transform(decoder, opts, outputImg)
	if err != nil {
		fmt.Printf("error transforming image, %s\n", err)
		os.Exit(1)
	}

	if outputFilename == "" {
		outputFilename = "resize" + filepath.Ext("inputFilename")
	}

	if _, err := os.Stat(outputFilename); !os.IsNotExist(err) {
		fmt.Printf("output filename %s exists, quitting\n", outputFilename)
		os.Exit(1)
	}

	err = ioutil.WriteFile(outputFilename, outputImg, 0400)
	if err != nil {
		fmt.Printf("error writing out resized image, %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("image written to %s\n", outputFilename)

  }
