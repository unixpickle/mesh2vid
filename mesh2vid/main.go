package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/ffmpego"
	"github.com/unixpickle/mesh2vid"
	"github.com/unixpickle/model3d/model3d"
)

func main() {
	var delta float64
	var fps float64
	var zStride int
	flag.Float64Var(&delta, "delta", 0.04, "spatial quantization")
	flag.Float64Var(&fps, "fps", 24.0, "frames per second in video")
	flag.IntVar(&zStride, "z-stride", 1, "stride along Z axis (to reduce Z resolution)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: mesh2vid [flags] <input.stl> <output.mp4>")
		fmt.Fprintln(os.Stderr)
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()

	if len(flag.Args()) != 2 {
		flag.Usage()
	}
	inFile := flag.Args()[0]
	outFile := flag.Args()[1]

	log.Println("Loading mesh...")
	r, err := os.Open(inFile)
	essentials.Must(err)
	defer r.Close()
	triangles, err := model3d.ReadSTL(r)
	essentials.Must(err)
	mesh := model3d.NewMeshTriangles(triangles)

	log.Println("Creating slicer...")
	slicer := mesh2vid.NewSlicer(mesh, delta)

	log.Println("Opening output video...")
	vw, err := ffmpego.NewVideoWriter(outFile, slicer.Width(), slicer.Height(), fps)
	essentials.Must(err)
	defer vw.Close()

	for i := 0; i < slicer.NumFrames()/zStride; i++ {
		log.Printf("Encoding frame %d/%d", (i + 1), slicer.NumFrames()/zStride)
		essentials.Must(vw.WriteFrame(slicer.Slice(i * zStride)))
	}
}
