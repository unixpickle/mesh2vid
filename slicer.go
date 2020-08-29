package mesh2vid

import (
	"image"
	"image/color"
	"math"

	"github.com/unixpickle/model3d/model3d"
)

// A Slicer produces cross-sections of a 3D model.
type Slicer struct {
	solid  model3d.Solid
	min    model3d.Coord3D
	max    model3d.Coord3D
	width  int
	height int
	frames int
	delta  float64
}

// NewSlicer creates a Slicer from a mesh.
//
// The resulting slicer will divide up space by a given
// delta, and produces slices along the Z axis.
func NewSlicer(mesh *model3d.Mesh, delta float64) *Slicer {
	min, max := mesh.Min(), mesh.Max()
	size := max.Sub(min).Scale(1 / delta)
	width := int(math.Ceil(size.X))
	height := int(math.Ceil(size.Y))

	// Force even dimensions to support video formats.
	if width%2 == 1 {
		width += 1
		min.X -= delta / 2
		max.X += delta / 2
	}
	if height%2 == 1 {
		height += 1
		min.Y -= delta / 2
		max.Y += delta / 2
	}

	frames := int(math.Ceil(size.Z))
	return &Slicer{
		solid:  model3d.NewColliderSolid(model3d.MeshToCollider(mesh)),
		min:    min,
		max:    max,
		width:  width,
		height: height,
		frames: frames,
		delta:  delta,
	}
}

// NumFrames gets the total number of frames that the
// slicer can produce for the model.
func (s *Slicer) NumFrames() int {
	return s.frames
}

// Width gets the width of cross-sections.
func (s *Slicer) Width() int {
	return s.width
}

// Height gets the height of cross-sections.
func (s *Slicer) Height() int {
	return s.height
}

// Slice gets a cross-section at the given frame index.
func (s *Slicer) Slice(frameIdx int) image.Image {
	if frameIdx < 0 || frameIdx >= s.NumFrames() {
		panic("frame index out of range")
	}
	z := s.min.Z + float64(frameIdx)*s.delta
	res := image.NewGray(image.Rect(0, 0, s.width, s.height))
	for y := 0; y < s.height; y++ {
		for x := 0; x < s.width; x++ {
			coord := model3d.XYZ(
				float64(x)*s.delta+s.min.X,
				float64(y)*s.delta+s.min.Y,
				z,
			)
			if !s.solid.Contains(coord) {
				res.SetGray(x, y, color.Gray{Y: 0xff})
			}
		}
	}
	return res
}
