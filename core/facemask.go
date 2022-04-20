// Copyright 2022 PJ Engineering and Business Solutions Pty. Ltd. All rights reserved.

package showerglass

import (
	_ "embed"
	"image"
	"io"
	"sort"
	"sync"

	"golang.org/x/image/draw"
	"golang.org/x/sync/errgroup"

	"github.com/esimov/caire"
	pigo "github.com/esimov/pigo/core"
	"github.com/esimov/triangle/v2"
)

//go:embed cascade/facefinder
var cascadeFile []byte

// ResizeAlg defines the algorithm to use for resizing purposes.
type ResizeAlg int

const (
	// Caire uses a content-aware image resizing algorithm (slowest, but more sophisticated results).
	// See: https://github.com/esimov/caire
	//
	// This is the default.
	Caire ResizeAlg = iota

	// NearestNeighbor is very fast (but very low quality results).
	// See: https://pkg.go.dev/golang.org/x/image/draw#pkg-variables
	NearestNeighbor

	// ApproxBiLinear is fast (but medium quality results).
	// See: https://pkg.go.dev/golang.org/x/image/draw#pkg-variables
	ApproxBiLinear

	// BiLinear is slow (but high quality results).
	// See: https://pkg.go.dev/golang.org/x/image/draw#pkg-variables
	BiLinear

	// CatmullRom is slower (but very high quality results).
	// See: https://pkg.go.dev/golang.org/x/image/draw#pkg-variables
	CatmullRom
)

// Options is used to adjust behavior of the FaceMask function.
type Options struct {

	// NewHeight, when set, is the height of the image after the resize process.
	// If a float64 is provided, NewHeight is interpreted as a percentage of the original image.
	// An int can also be provided to mean the actual NewHeight.
	// A value of 0 indicates no change from the original height.
	NewHeight interface{}

	// NewWidth, when set, is the width of the image after the resize process.
	// If a float64 is provided, NewWidth is interpreted as a percentage of the original image.
	// An int can also be provided to mean the actual NewWidth.
	// A value of 0 indicates no change from the original width.
	NewWidth interface{}

	// TriangleConfig is called for each detected face. A *Processor must be returned to
	// modify the delaunay triangulation algorithm parameters.
	// If nil is returned, no triangulation is performed.
	//
	// facearea is the size of the detected face. You can calibrate the returned *Processor parameters based on
	// the area. The 2 most relevant parameters are BlurRadius and MaxPoints.
	//
	// QRank indicates the rank of all the "detected faces". If only one face is expected,
	// then you can return nil for all QRank > 0. If the Q value is _sufficiently_ low, you can
	// presume a false positive and also return nil.
	//
	// See: https://pkg.go.dev/github.com/esimov/triangle#Processor
	TriangleConfig func(QRank, facearea int, Q float32, h, w int) *Processor

	// ResizeAlg sets which resizing algorithm to use.
	// The default is "Caire".
	ResizeAlg ResizeAlg
}

// Processor is a triangle.Processor.
//
// See: https://pkg.go.dev/github.com/esimov/triangle#Processor
type Processor = triangle.Processor

// FaceMask accepts an io.Reader (usually an *os.File) and returns an image with the FaceMask filter applied.
// The type of image format used for the input is also returned.
func FaceMask(input io.Reader, opts ...Options) (image.Image, string, error) {
	_src, format, err := image.Decode(input)
	if err != nil {
		return nil, format, err
	}

	src := pigo.ImgToNRGBA(_src)

	var (
		oh int = src.Bounds().Max.Y - src.Bounds().Min.Y
		ow int = src.Bounds().Max.X - src.Bounds().Min.X

		nh int = oh
		nw int = ow
	)

	if len(opts) > 0 {
		if opts[0].NewHeight != nil {
			switch v := opts[0].NewHeight.(type) {
			case int:
				if v != 0 {
					nh = v
				}
			case float64:
				if v != 0.0 {
					nh = int(float64(oh) * v / 100.0)
				}
			default:
				panic("NewHeight must be an int or float64")
			}
		}

		if opts[0].NewWidth != nil {
			switch v := opts[0].NewWidth.(type) {
			case int:
				if v != 0 {
					nw = v
				}
			case float64:
				if v != 0.0 {
					nw = int(float64(ow) * v / 100.0)
				}
			default:
				panic("NewWidth must be an int or float64")
			}
		}
	}

	// Step 1: Resize image
	var resized image.Image = src
	if nh != oh || nw != ow {
		p := &caire.Processor{
			NewWidth:   nw,
			NewHeight:  nh,
			FaceDetect: true,
		}

		if len(opts) > 0 {
			if opts[0].ResizeAlg == Caire {
				// Run caire
				var err error
				resized, err = p.Resize(src)
				if err != nil {
					return nil, format, err
				}
			}
		} else {
			// Run caire (default)
			var err error
			resized, err = p.Resize(src)
			if err != nil {
				return nil, format, err
			}
		}
	}

	// Step 2: Search for faces
	pixels := pigo.RgbToGrayscale(resized)
	cols, rows := resized.Bounds().Max.X, resized.Bounds().Max.Y

	cParams := pigo.CascadeParams{
		MinSize:     20,
		MaxSize:     1000,
		ShiftFactor: 0.1,
		ScaleFactor: 1.1,

		ImageParams: pigo.ImageParams{
			Pixels: pixels,
			Rows:   rows,
			Cols:   cols,
			Dim:    cols,
		},
	}

	pigo := pigo.NewPigo()
	classifier, err := pigo.Unpack(cascadeFile)
	if err != nil {
		return nil, format, err
	}
	dets := classifier.RunCascade(cParams, 0.0)
	dets = classifier.ClusterDetections(dets, 0.2)

	if len(dets) == 0 {
		if len(opts) > 0 && (nh != oh || nw != ow) && opts[0].ResizeAlg != Caire {
			resized, _ = Resize(resized, nh, nw, opts[0].ResizeAlg)
		}
		return resized, format, nil
	}

	// Step 3: Rank which "detected faces" are best.
	sort.Slice(dets, func(i, j int) bool { return dets[i].Q > dets[j].Q })

	// Create triangle template which is totally black, except for location of faces.
	// Those locations will contain triangulated versions of the faces.
	g := new(errgroup.Group)
	lock := sync.Mutex{}
	trigFaceTemplate := image.NewNRGBA(resized.Bounds())

	// Create a Union mask
	unionLock := sync.Mutex{}
	unionMask := image.NewNRGBA(resized.Bounds())

	for idx, det := range dets {
		det := det
		idx := idx
		g.Go(func() error {
			facesize := image.Rectangle{
				Min: image.Point{
					X: int(float64(det.Col) - float64(det.Scale)/2.0),
					Y: int(float64(det.Row) - float64(det.Scale)/2.0),
				},
				Max: image.Point{
					X: int(float64(det.Col) + float64(det.Scale)/2.0),
					Y: int(float64(det.Row) + float64(det.Scale)/2.0),
				},
			}

			area := (facesize.Bounds().Max.Y - facesize.Bounds().Min.Y) * (facesize.Bounds().Max.X - facesize.Bounds().Min.X)
			var tp *Processor
			if len(opts) > 0 && opts[0].TriangleConfig != nil {
				tp = opts[0].TriangleConfig(idx, area, det.Q, resized.Bounds().Dy(), resized.Bounds().Dx())
				if tp == nil {
					return nil
				}
			} else {
				tp = &Processor{}
			}

			// Add to union mask
			ellipse := &ellipse{
				cx:     det.Col,
				cy:     det.Row,
				rx:     int(float64(det.Scale) * 0.8 / 2),
				ry:     int(float64(det.Scale) * 0.8 / 1.6),
				width:  resized.Bounds().Dx(),
				height: resized.Bounds().Dy(),
			}
			unionLock.Lock()
			draw.Draw(unionMask, unionMask.Bounds(), ellipse, image.Point{}, draw.Over)
			unionLock.Unlock()

			// Extract from resized just the portion that is contained in rect
			new := image.NewNRGBA(facesize.Bounds())
			draw.Draw(new, new.Bounds(), resized, new.Bounds().Min, draw.Over)

			// Step 4: Run Triangle algorithm
			img := &triangle.Image{*tp}
			triangled, _, _, err := img.Draw(new, *tp, func() {})
			if err != nil {
				return err
			}

			// Paste triangled image into trigFaceTemplate
			lock.Lock()
			draw.Draw(trigFaceTemplate, facesize.Bounds(), triangled, image.Point{}, draw.Over)
			lock.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, format, err
	}

	// Step 5: Draw Triangulated faces on top of original (resized) image
	draw.DrawMask(resized.(*image.NRGBA), resized.Bounds(), trigFaceTemplate, image.Point{}, unionMask, image.Point{}, draw.Over)

	// Step 6: Final resize
	if len(opts) > 0 && (nh != oh || nw != ow) && opts[0].ResizeAlg != Caire {
		resized, _ = Resize(resized, nh, nw, opts[0].ResizeAlg)
	}

	return resized, format, nil
}
