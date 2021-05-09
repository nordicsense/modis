package dataset

import (
	"fmt"
	"math"
	"time"

	"github.com/arcticgeo/modis"
)

func NewInMemory(p *modis.ImageParams) *inMemory {
	data := make([][]float64, p.YSize()) // FIXME: by row?
	for i := range data {
		data[i] = make([]float64, p.XSize())
		if i == 0 {
			for j := 0; j < p.XSize(); j++ {
				data[i][j] = math.NaN()
			}
		} else {
			copy(data[i], data[0])
		}
	}
	return &inMemory{data: data, p: p}
}

type inMemory struct {
	data [][]float64
	p    *modis.ImageParams
}

func (ds *inMemory) ImageParams() *modis.ImageParams {
	return ds.p
}

func (ds *inMemory) Read(x, y int) (float64, error) {
	if res, err := ds.ReadBlock(x, y, modis.Box{0, 0, 1, 1}); err == nil {
		return res[0], nil
	} else {
		return math.NaN(), err
	}
}

func (ds *inMemory) ReadAtLatLon(ll modis.LatLon) (float64, error) {
	x, y := ds.ImageParams().Transform().LatLon2Pixels(ll)
	return ds.Read(x, y)
}

func (ds *inMemory) ReadTime(x, y int) (time.Time, error) {
	v, err := ds.Read(x, y)
	if err != nil {
		return time.Time{}, err
	}
	ll := ds.ImageParams().Transform().Pixels2LatLon(x, y)
	return ds.ImageParams().Value2time(v, ll)
}

func (ds *inMemory) ReadTimeAtLatLon(ll modis.LatLon) (time.Time, error) {
	v, err := ds.ReadAtLatLon(ll)
	if err != nil {
		return time.Time{}, err
	}
	return ds.ImageParams().Value2time(v, ll)
}

func (ds *inMemory) ReadBlock(x, y int, box modis.Box) ([]float64, error) {
	buffer := make([]float64, box[2]*box[3])
	for j := 0; j < box[3]; j++ {
		for i := 0; i < box[2]; i++ {
			buffer[j*box[2]+i] = ds.data[y+box[1]+j][x+box[0]+i]
		}
	}
	return buffer, nil
}

func (ds *inMemory) Write(x, y int, v float64) error {
	return ds.WriteBlock(x, y, modis.Box{0, 0, 1, 1}, []float64{v})
}

func (ds *inMemory) WriteAtLatLon(ll modis.LatLon, v float64) error {
	x, y := ds.ImageParams().Transform().LatLon2Pixels(ll)
	return ds.Write(x, y, v)
}

func (ds *inMemory) WriteBlock(x, y int, box modis.Box, buffer []float64) error {
	for j := 0; j < box[3]; j++ {
		for i := 0; i < box[2]; i++ {
			ds.data[y+box[1]+j][x+box[0]+i] = buffer[j*box[2]+i]
		}
	}
	return nil
}

func (ds *inMemory) ToFileWriter(fileName string, driver Driver) (Writer, error) {
	// FIXME - implement
	return nil, fmt.Errorf("not implemented")
}
