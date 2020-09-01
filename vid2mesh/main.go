package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/ffmpego"
	"github.com/unixpickle/mesh2vid"
	"github.com/unixpickle/model3d/model3d"
)

func main() {
	var delta float64
	var mcDelta float64
	var zStride int
	var smoothIters int
	var maxFrames int
	flag.Float64Var(&delta, "delta", 0.04, "spatial quantization")
	flag.Float64Var(&mcDelta, "mc-delta", 0,
		"delta to use for maching cubes, if different than the spatial delta above")
	flag.IntVar(&zStride, "z-stride", 1, "stride along Z axis (to reduce Z resolution)")
	flag.IntVar(&smoothIters, "smooth", 20, "smoothing iterations to apply to each frame")
	flag.IntVar(&maxFrames, "max-frames", 0, "maximum number of frames to read")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: vid2mesh [flags] <input.mp4> <output.stl>")
		fmt.Fprintln(os.Stderr)
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()
	if mcDelta == 0 {
		mcDelta = delta
	}

	if len(flag.Args()) != 2 {
		flag.Usage()
	}
	inFile := flag.Args()[0]
	outFile := flag.Args()[1]

	deslicer := mesh2vid.NewDeslicer(delta, zStride, smoothIters)

	log.Println("Reading slices from video...")
	vr, err := ffmpego.NewVideoReader(inFile)
	essentials.Must(err)
	defer vr.Close()
	for i := 0; true; i++ {
		frame, err := vr.ReadFrame()
		if err == io.EOF {
			break
		}
		essentials.Must(err)
		log.Printf("Processing frame %d...", i+1)
		deslicer.AddFrame(frame)
		if maxFrames > 0 && i+1 >= maxFrames {
			break
		}
	}

	log.Println("Converting sliced solid to mesh ...")
	mesh := model3d.MarchingCubesSearch(deslicer, mcDelta, 8)
	log.Printf("Saving to %s...", outFile)
	mesh.SaveGroupedSTL(outFile)
}
