package mesh2vid

import (
	"image"
	"image/color"
	"math"

	"github.com/unixpickle/model3d/model3d"
)

// A Slicer produces cross-sections of a 3D model.
type Slicer struct {
	solid    model3d.Solid
	collider model3d.RectCollider
	min      model3d.Coord3D
	max      model3d.Coord3D
	width    int
	height   int
	frames   int
	delta    float64
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
	collider := model3d.MeshToCollider(mesh)
	return &Slicer{
		solid:    model3d.NewColliderSolid(collider),
		collider: model3d.MeshToCollider(mesh),
		min:      min,
		max:      max,
		width:    width,
		height:   height,
		frames:   frames,
		delta:    delta,
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
	z := s.min.Z + (float64(frameIdx)+0.5)*s.delta
	res := image.NewGray(image.Rect(0, 0, s.width, s.height))
	s.sliceRange(res, z, 0, 0, s.width, s.height)
	return res
}

func (s *Slicer) sliceRange(img *image.Gray, z float64, x, y, width, height int) {
	if width == 0 || height == 0 {
		return
	}

	if width == 1 && height == 1 {
		s.sliceRangeConstant(img, z, x, y, width, height)
		return
	}

	rect := &model3d.Rect{
		MinVal: model3d.XYZ(
			float64(x)*s.delta+s.min.X,
			float64(y)*s.delta+s.min.Y,
			z-s.delta/2,
		),
		MaxVal: model3d.XYZ(
			(float64(x)+float64(width))*s.delta+s.min.X,
			(float64(y)+float64(height))*s.delta+s.min.Y,
			z+s.delta/2,
		),
	}
	if !s.collider.RectCollision(rect) {
		// If the entire bounds do not touch the surface, then
		// we only need to do one containment check.
		s.sliceRangeConstant(img, z, x, y, width, height)
		return
	}

	for hIdx := 0; hIdx < 2; hIdx++ {
		var subY, subHeight int
		if hIdx == 0 {
			subY = y
			subHeight = height / 2
		} else {
			subY = y + height/2
			subHeight = height - (height / 2)
		}
		for wIdx := 0; wIdx < 2; wIdx++ {
			var subX, subWidth int
			if wIdx == 0 {
				subX = x
				subWidth = width / 2
			} else {
				subX = x + width/2
				subWidth = width - (width / 2)
			}

			s.sliceRange(img, z, subX, subY, subWidth, subHeight)
		}
	}
}

func (s *Slicer) sliceRangeConstant(img *image.Gray, z float64, x, y, width, height int) {
	coord := model3d.XYZ(
		(float64(x)+float64(width)/2.0)*s.delta+s.min.X,
		(float64(y)+float64(height)/2.0)*s.delta+s.min.Y,
		z,
	)
	if !s.solid.Contains(coord) {
		for subY := 0; subY < height; subY++ {
			for subX := 0; subX < width; subX++ {
				img.SetGray(x+subX, y+subY, color.Gray{Y: 0xff})
			}
		}
	}
}
