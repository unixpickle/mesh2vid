package mesh2vid

import (
	"image"
	"image/color"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

// A Deslicer is a Solid that interpolates between 2D
// slices.
type Deslicer struct {
	delta       float64
	zDelta      float64
	smoothIters int

	layers []model2d.SDF
	min    model3d.Coord3D
	max    model3d.Coord3D
}

func NewDeslicer(delta float64, stride, smoothIters int) *Deslicer {
	return &Deslicer{
		delta:       delta,
		zDelta:      delta * float64(stride),
		smoothIters: smoothIters,

		min: model3d.Coord3D{},
		max: model3d.Coord3D{},
	}
}

// AddFrame adds a layer to the model on top of the
// previous layers on the Z axis.
func (d *Deslicer) AddFrame(frame image.Image) {
	mesh := model2d.NewBitmapImage(frame, func(c color.Color) bool {
		r, _, _, _ := c.RGBA()
		return r < 0xffff/2
	}).Mesh().SmoothSq(d.smoothIters).Scale(d.delta)
	sdf := model2d.MeshToSDF(mesh)
	d.layers = append(d.layers, sdf)
	d.max.Z = float64(len(d.layers)-1) * d.zDelta

	min, max := mesh.Min(), mesh.Max()
	d.max.X = math.Max(d.max.X, max.X)
	d.max.Y = math.Max(d.max.Y, max.Y)
	d.min.X = math.Min(d.min.X, min.X)
	d.min.Y = math.Min(d.min.Y, min.Y)
}

// Min gets the current minimum coordinate in the solid.
func (d *Deslicer) Min() model3d.Coord3D {
	return d.min
}

// Max gets the current maximum coordinate in the solid.
func (d *Deslicer) Max() model3d.Coord3D {
	return d.max
}

func (d *Deslicer) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(d, c) || len(d.layers) == 0 {
		return false
	}
	z1 := int(c.Z / d.zDelta)
	z2 := z1 + 1
	s1 := d.getLayerSDF(z1, c)
	s2 := d.getLayerSDF(z2, c)
	fracInZ2 := (c.Z - float64(z1)*d.zDelta) / d.zDelta
	sdf := s1*(1-fracInZ2) + s2*fracInZ2
	return sdf > 0
}

func (d *Deslicer) getLayerSDF(layerIdx int, c model3d.Coord3D) float64 {
	if layerIdx < 0 {
		layerIdx = 0
	} else if layerIdx >= len(d.layers) {
		layerIdx = len(d.layers) - 1
	}
	return d.layers[layerIdx].SDF(c.XY())
}
