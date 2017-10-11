// Lists the streams and some codec details of a media file
//
// Tested with
//
// $ go run examples/mediainfo/medianfo.go --input=https://bintray.com/imkira/go-libav/download_file?file_path=sample_iPod.m4v
//
// stream 0: eng aac audio, 2 channels, 44100 Hz
// stream 1: h264 video, 320x240

package main

//#include <libavutil/avutil.h>
//#include <libavutil/avstring.h>
//#include <libavcodec/avcodec.h>
//#include <libavformat/avformat.h>
//#include <libavformat/avio.h>
//
// #cgo pkg-config: libavformat libavutil
import "C"

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"unsafe"

	"github.com/imkira/go-libav/avcodec"
	"github.com/imkira/go-libav/avformat"
	"github.com/imkira/go-libav/avutil"
)

var inputFileName string

func init() {
	flag.StringVar(&inputFileName, "input", "", "source file to probe")
	flag.Parse()
}

func main() {
	if len(inputFileName) == 0 {
		log.Fatalf("Missing --input=file\n")
	}

	// open format (container) context
	decFmt, err := avformat.NewContextForInput()
	if err != nil {
		log.Fatalf("Failed to open input context: %v", err)
	}

	// set some options for opening file
	options := avutil.NewDictionary()
	defer options.Free()

	// mods
	//var cCtx *C.AVIOContext
	//var buffer []byte

	buffer, err := ioutil.ReadFile(inputFileName)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(buffer))

	var bufferSize C.int
	bufferSize = C.int(8192)
	readBufferSize := C.size_t(bufferSize)
	readExchangeArea := C.av_malloc(readBufferSize)

	var readFunction *[0]byte

	cCtx := C.avio_alloc_context((*C.uchar)(readExchangeArea), bufferSize, 0, nil, readFunction, nil, nil)
	ioCtx := avformat.NewIOContextFromC(unsafe.Pointer(cCtx))
	decFmt.SetIOContext(ioCtx)
	// mods

	// open file for decoding
	if err := decFmt.OpenInput("", nil, options); err == nil {
		log.Fatalf("Failed to open input file: %v", err)
	}
	defer decFmt.CloseInput()

	// initialize context with stream information
	if err := decFmt.FindStreamInfo(nil); err != nil {
		log.Fatalf("Failed to find stream info: %v", err)
	}

	// show stream info
	for _, stream := range decFmt.Streams() {
		language := stream.MetaData().Get("language")
		streamCtx := stream.CodecContext()
		codecID := streamCtx.CodecID()
		descriptor := avcodec.CodecDescriptorByID(codecID)
		switch streamCtx.CodecType() {
		case avutil.MediaTypeVideo:
			width := streamCtx.Width()
			height := streamCtx.Height()
			fmt.Printf("stream %d: %s video, %dx%d\n",
				stream.Index(),
				descriptor.Name(),
				width,
				height)
		case avutil.MediaTypeAudio:
			channels := streamCtx.Channels()
			sampleRate := streamCtx.SampleRate()
			fmt.Printf("stream %d: %s %s audio, %d channels, %d Hz\n",
				stream.Index(),
				language,
				descriptor.Name(),
				channels,
				sampleRate)
		case avutil.MediaTypeSubtitle:
			fmt.Printf("stream %d: %s %s subtitle\n",
				stream.Index(),
				language,
				descriptor.Name())
		}
	}
}
