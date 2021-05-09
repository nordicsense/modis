package dataset

import (
	"time"

	"github.com/arcticgeo/modis"
)

type Reader interface {
	ImageParams() *modis.ImageParams
	Read(x, y int) (float64, error)
	ReadAtLatLon(ll modis.LatLon) (float64, error)
	ReadTime(x, y int) (time.Time, error)
	ReadTimeAtLatLon(ll modis.LatLon) (time.Time, error)
	ReadBlock(x, y int, box modis.Box) ([]float64, error)
	ToMemory() *inMemory
	Close()
}

type Writer interface {
	ImageParams() *modis.ImageParams
	Write(x, y int, v float64) error
	WriteAtLatLon(ll modis.LatLon, v float64) error
	WriteBlock(x, y int, box modis.Box, buffer []float64) error
	Close()
}
