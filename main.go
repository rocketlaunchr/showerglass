package main

import (
	"image/jpeg"
	_ "image/png"
	"log"
	"os"

	"github.com/rocketlaunchr/showerglass/core"
)

func main() {

	f, err := os.Open("face.jpg")
	if err != nil {
		log.Fatalf("Error opening the file: %v", err)
	}
	defer f.Close()

	opts := showerglass.Options{
		NewHeight: 100.0,
		NewWidth:  100.0,
		ResizeAlg: showerglass.CatmullRom,
		TriangleConfig: func(QRank, facearea int, Q float32, h, w int) *showerglass.Processor {
			if QRank < 1 {
				// only modify first detected face
				return &showerglass.Processor{
					MaxPoints:  4000,
					BlurRadius: 4,
					BlurFactor: 1,
					EdgeFactor: 6,
					PointRate:  0.075,
				}
			}
			return nil
		},
	}

	filtered, _, err := showerglass.FaceMask(f, opts)
	if err != nil {
		log.Fatalf("Error aplying filter: %v", err)
	}

	out, err := os.Create("facemask.jpg")
	if err != nil {
		log.Fatalf("Error writing the file: %s", err)
	}

	err = jpeg.Encode(out, filtered, &jpeg.Options{Quality: 100})
	if err != nil {
		log.Fatalf("Error encoding to file: %s", err)
	}

}
