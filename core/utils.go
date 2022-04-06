// Copyright 2022 PJ Engineering and Business Solutions Pty. Ltd. All rights reserved.

package showerglass

import (
	"golang.org/x/image/draw"
	"image"

	"github.com/esimov/caire"
)

func copyImage(src image.Image) *image.NRGBA {
	b := src.Bounds()
	m := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(m, m.Bounds(), src, b.Min, draw.Src)
	return m
}

// Resize will resize src to a new height (nh) and new width (nw).
// For the Caire algorithm (default), src must be an *image.NRGBA.
//
// See: https://pkg.go.dev/github.com/rocketlaunchr/showerglass/core#ResizeAlg
func Resize(src image.Image, nh, nw int, alg ResizeAlg) (image.Image, error) {
	dst := image.NewNRGBA(image.Rect(0, 0, nw, nh))
	if alg == NearestNeighbor {
		draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	} else if alg == ApproxBiLinear {
		draw.ApproxBiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	} else if alg == BiLinear {
		draw.BiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	} else if alg == CatmullRom {
		draw.CatmullRom.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	} else {
		p := &caire.Processor{
			NewWidth:   nw,
			NewHeight:  nh,
			FaceDetect: true,
		}
		resized, err := p.Resize(src.(*image.NRGBA))
		if err != nil {
			return nil, err
		}
		return resized, nil
	}
	return dst, nil
}

// ConvertToGrayscale will convert an image from color to grayscale.
func ConvertToGrayscale(src image.Image) image.Image {
	result := image.NewGray(src.Bounds())
	draw.Draw(result, result.Bounds(), src, src.Bounds().Min, draw.Src)
	return result
}
