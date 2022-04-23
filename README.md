<p align="right">
  ⭐ &nbsp;&nbsp;<strong>the project to show your appreciation.</strong> :arrow_upper_right:
</p>

<p align="right">
  <a href="http://godoc.org/github.com/rocketlaunchr/showerglass/core"><img src="http://godoc.org/github.com/rocketlaunchr/showerglass/core?status.svg" /></a>
  <a href="https://goreportcard.com/report/github.com/rocketlaunchr/showerglass/core"><img src="https://goreportcard.com/badge/github.com/rocketlaunchr/showerglass/core" /></a>
  <a href="https://gocover.io/github.com/rocketlaunchr/showerglass/core"><img src="http://gocover.io/_badge/github.com/rocketlaunchr/showerglass/core" /></a>
</p>


# Showerglass

A soothing face filter where you can appreciate the beauty but not fully identify the person.
Useful for social applications, blogging etc.

You can read how it works in this [article](https://itnext.io/profile-photos-privacy-and-social-media-e66a908cd054). 

# Features

1. Resizing (Caire, NearestNeighbor, ApproxBiLinear, BiLinear, CatmullRom)
2. Automatic Face detection
3. _Frosted_ Showerglass filter (delaunay triangulation) over only face

<p align="center">
<img src="https://github.com/rocketlaunchr/showerglass/raw/master/example.jpg" alt="female face" />
</p>

Image credit: https://unsplash.com/photos/tCJ44OIqceU

## Installation

```
go get -u github.com/rocketlaunchr/showerglass/core
```

```go
import "github.com/rocketlaunchr/showerglass/core"
```

## Usage


```go
import	("image/jpeg"; "os";)
import	"github.com/rocketlaunchr/showerglass/core"

f, _ := os.Open("face.jpg")
defer f.Close()

opts := showerglass.Options{
	NewHeight: 100.0,
	NewWidth:  100.0,
	ResizeAlg: showerglass.CatmullRom,
	TriangleConfig: func(QRank, facearea int, Q float32, h, w int, c showerglass.MaxPoints) *showerglass.TriangleConfig {
		if QRank < 1 {
			// only modify first detected face
			return &showerglass.TriangleConfig{
				MaxPoints:  1500,
				BlurRadius: 4,
				BlurFactor: 1,
				EdgeFactor: 6,
				PointRate:  0.075,
			}
		}
		return nil
	},
}

masked, _, _ := showerglass.FaceMask(f, opts)

out, _ := os.Create("masked.jpg")

jpeg.Encode(out, masked, &jpeg.Options{Quality: 100})
```

### Calibration

* A higher `MaxPoints` means the face looks closer to the original.
* A lower `MaxPoints` (with the exception of `0`) means a more obfuscated face.

Based on the `facearea`, you need to calibrate `MaxPoints` to achieve the desired _feel_.

## Credits

- [Endre Simo](https://github.com/esimov) - One of the masters of Image Processing **[worth following]**


## Other useful packages

- [awesome-svelte](https://github.com/rocketlaunchr/awesome-svelte) - Resources for killing react
- [dataframe-go](https://github.com/rocketlaunchr/dataframe-go) - Statistics and data manipulation
- [dbq](https://github.com/rocketlaunchr/dbq) - Zero boilerplate database operations for Go
- [electron-alert](https://github.com/rocketlaunchr/electron-alert) - SweetAlert2 for Electron Applications
- [google-search](https://github.com/rocketlaunchr/google-search) - Scrape google search results
- [igo](https://github.com/rocketlaunchr/igo) - A Go transpiler with cool new syntax such as fordefer (defer for for-loops)
- [mysql-go](https://github.com/rocketlaunchr/mysql-go) - Properly cancel slow MySQL queries
- [react](https://github.com/rocketlaunchr/react) - Build front end applications using Go
- [remember-go](https://github.com/rocketlaunchr/remember-go) - Cache slow database queries
- [testing-go](https://github.com/rocketlaunchr/testing-go) - Testing framework for unit testing

#

### Legal Information

The license is a modified MIT license. Refer to `LICENSE` file for more details.

**© 2022 PJ Engineering and Business Solutions Pty. Ltd.**