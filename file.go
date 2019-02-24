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
  "context"
  "github.com/aws/aws-lambda-go/events"
)

var EncodeOptions = map[string]map[int]int{
	".jpeg": map[int]int{lilliput.JpegQuality: 85},
	".png":  map[int]int{lilliput.PngCompression: 7},
	".webp": map[int]int{lilliput.WebpQuality: 85},
}

type S3Resp struct {
	Records []struct {
		Cf struct {
			Config struct {
				DistributionDomainName string `json:"distributionDomainName"`
				DistributionID         string `json:"distributionId"`
				EventType              string `json:"eventType"`
				RequestID              string `json:"requestId"`
			} `json:"config"`
			Request struct {
				Body struct {
					Action         string `json:"action"`
					Data           string `json:"data"`
					Encoding       string `json:"encoding"`
					InputTruncated bool   `json:"inputTruncated"`
				} `json:"body"`
				ClientIP    string `json:"clientIp"`
				Querystring string `json:"querystring"`
				URI         string `json:"uri"`
				Method      string `json:"method"`
				Headers     struct {
					Host []struct {
						Key   string `json:"key"`
						Value string `json:"value"`
					} `json:"host"`
					UserAgent []struct {
						Key   string `json:"key"`
						Value string `json:"value"`
					} `json:"user-agent"`
				} `json:"headers"`
				Origin struct {
					Custom struct {
						CustomHeaders struct {
							MyOriginCustomHeader []struct {
								Key   string `json:"key"`
								Value string `json:"value"`
							} `json:"my-origin-custom-header"`
						} `json:"customHeaders"`
						DomainName       string   `json:"domainName"`
						KeepaliveTimeout int      `json:"keepaliveTimeout"`
						Path             string   `json:"path"`
						Port             int      `json:"port"`
						Protocol         string   `json:"protocol"`
						ReadTimeout      int      `json:"readTimeout"`
						SslProtocols     []string `json:"sslProtocols"`
					} `json:"custom"`
					S3 struct {
						AuthMethod    string `json:"authMethod"`
						CustomHeaders struct {
							MyOriginCustomHeader []struct {
								Key   string `json:"key"`
								Value string `json:"value"`
							} `json:"my-origin-custom-header"`
						} `json:"customHeaders"`
						DomainName string `json:"domainName"`
						Path       string `json:"path"`
						Region     string `json:"region"`
					} `json:"s3"`
				} `json:"origin"`
			} `json:"request"`
		} `json:"cf"`
	} `json:"Records"`
}

func HandleRequest(ctx context.Context, s3Event events.S3Event) {
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

  func main() {
	lambda.Start(HandleRequest)
  }