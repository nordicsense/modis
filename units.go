package modis

import (
	"errors"
	"fmt"
	"math"

	"github.com/nordicsense/gdal"
)

const ModisWKT = `PROJCS["MODIS",
    GEOGCS["Unknown datum based upon the custom spheroid",
        DATUM["Not specified (based on custom spheroid)",
            SPHEROID["Custom spheroid",6371007.181,0]],
        PRIMEM["Greenwich",0],
        UNIT["degree",0.0174532925199433]],
    PROJECTION["Sinusoidal"],
    PARAMETER["longitude_of_center",0],
    PARAMETER["false_easting",0],
    PARAMETER["false_northing",0],
    UNIT["Meter",1]]`

// LatLon represents a latitude/longitude pair.
type LatLon [2]float64

// Box defines an area of raster: x,y offset and x,y size.
type Box [4]int

// Transform coordinates from one ESPG projection into another.
func (ll LatLon) Transform(fromESPG, toESPG int) (LatLon, error) {
	from, err := ll.CSRFromESPG(fromESPG)
	if err != nil {
		return ll, err
	}
	defer from.Destroy()
	to, err := ll.CSRFromESPG(toESPG)
	if err != nil {
		return ll, err
	}
	defer to.Destroy()
	return ll.transform(from, to)
}

// Transform coordinates from one ESPG projection into another.
func (ll LatLon) transform(from, to gdal.SpatialReference) (LatLon, error) {
	t := gdal.CreateCoordinateTransform(from, to)
	defer t.Destroy()
	lat := []float64{ll[0]}
	lon := []float64{ll[1]}
	z := []float64{0.0}
	if ok := t.Transform(1, lon, lat, z); ok {
		return LatLon{lat[0], lon[0]}, nil
	}
	return LatLon{lat[0], lon[0]}, errors.New("transformation failed")
}

func (ll LatLon) CSRFromESPG(espg int) (gdal.SpatialReference, error) {
	res := gdal.CreateSpatialReference("")
	err := res.FromEPSG(espg)
	return res, err
}

func (ll LatLon) MODIS_CSR() gdal.SpatialReference {
	return gdal.CreateSpatialReference(ModisWKT)
}

// Degrees2Sin transforms coordinates from the World Geodetic System (WGS84, given in degrees) into Sphere Sinusoidal.
func (ll LatLon) Degrees2Sin() (LatLon, error) {
	from, err := ll.CSRFromESPG(4326)
	if err != nil {
		return ll, err
	}
	to := ll.MODIS_CSR()
	defer from.Destroy()
	defer to.Destroy()
	return ll.transform(from, to)
}

// Sin2Degrees transforms coordinates from the Sphere Sinusoidal system into the World Geodetic System (WGS84).
func (ll LatLon) Sin2Degree() (LatLon, error) {
	from := ll.MODIS_CSR()
	defer from.Destroy()
	to, err := ll.CSRFromESPG(4326)
	if err != nil {
		return ll, err
	}
	defer to.Destroy()
	return ll.transform(from, to)
}

func (ll LatLon) String() string {
	return fmt.Sprintf("(%.2f,%.2f)", ll[0], ll[1])
}

// AffineTransform defines the transformation of the projection.
type AffineTransform [6]float64

// Performs the direct affine transform from image pixels to World Sinusoidal coordinates.
func (at AffineTransform) Pixels2LatLonSin(x, y int) LatLon {
	lat := float64(y)*at[5] + at[3]
	lon := float64(x)*at[1] + at[0]
	return LatLon{lat, lon}
}

// Performs the direct affine transform from image pixels to lat/lon in degrees.
func (at AffineTransform) Pixels2LatLon(x, y int) LatLon {
	lat := float64(y)*at[5] + at[3]
	lon := float64(x)*at[1] + at[0]
	res, _ := LatLon{lat, lon}.Sin2Degree()
	return res
}

// Performs the inverse affine transform from World Sinusoidal coordinates to image pixels.
func (at AffineTransform) LatLonSin2Pixels(ll LatLon) (int, int) {
	x := int(math.Round((ll[1] - at[0]) / at[1]))
	y := int(math.Round((ll[0] - at[3]) / at[5]))
	return x, y
}

// Performs the inverse affine transform from lat/lon in degrees to image pixels.
func (at AffineTransform) LatLon2Pixels(ll LatLon) (int, int) {
	ll, _ = ll.Degrees2Sin()
	return at.LatLonSin2Pixels(ll)
}

// ModisLST2UTC transforms MODIS time values in hours given from Local Solar Time to UTC.
func ModisLST2UTC(lst, lonDegree float64) float64 {
	offset := 0.0
	if lst < 0.0 {
		offset = 24.0
	} else if lst >= 24.0 {
		offset = -24.0 // FIXME should this be next day?
	}
	return lst - lonDegree/15.0 + offset
}
