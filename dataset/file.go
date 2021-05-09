package dataset

import (
	"fmt"
	"math"
	"time"

	"github.com/arcticgeo/modis"
	"github.com/lukeroth/gdal"
)

type Driver string

const (
	GTiff Driver = "GTiff"

	band   = 1
	bands  = 1
	domain = ""
)

func Open(fileName string) (Reader, error) {
	ds, err := gdal.Open(fileName, gdal.ReadOnly)
	if err != nil {
		return nil, err
	}
	if ds.RasterCount() < 1 {
		return nil, fmt.Errorf("no raster bands found")
	}
	rb := ds.RasterBand(band)
	dt, _ := time.Parse("2006-01-02", ds.MetadataItem("RANGEBEGINNINGDATE", "")) // ignore missing date
	b := modis.ImageParamsBuilder(ds.RasterXSize(), ds.RasterYSize()).
		DataType(rb.RasterDataType()).
		Transform(ds.GeoTransform()).
		Projection(ds.Projection()).
		Date(dt)
	if nan, ok := rb.NoDataValue(); ok {
		b = b.NaN(nan)
	}
	if scale, ok := rb.GetScale(); ok {
		b = b.Scale(scale)
	}
	if offset, ok := rb.GetOffset(); ok {
		b = b.Offset(offset)
	}
	for _, k := range ds.Metadata(domain) {
		b = b.Metadata(k, ds.MetadataItem(k, domain))
	}
	return &imageFile{Dataset: ds, p: b.Build()}, nil
}

func New(fileName string, driver Driver, p *modis.ImageParams) (Writer, error) {
	gdalDriver, err := gdal.GetDriverByName(string(driver))
	if err != nil {
		return nil, err
	}
	ds := gdalDriver.Create(fileName, p.XSize(), p.YSize(), bands, p.DataType(), nil)
	if err = ds.SetGeoTransform(p.Transform()); err != nil {
		return nil, err
	}
	if err = ds.SetProjection(p.Projection()); err != nil {
		return nil, err
	}
	rb := ds.RasterBand(band)
	if nan, ok := p.NaN(); ok {
		if err = rb.SetNoDataValue(nan); err != nil {
			return nil, err
		}
	}
	if err = rb.SetOffset(p.Offset()); err != nil {
		return nil, err
	}
	if err = rb.SetScale(p.Scale()); err != nil {
		return nil, err
	}
	for k, v := range p.Metadata() {
		if err := ds.SetMetadataItem(k, v, domain); err != nil {
			return nil, err
		}
	}
	return &imageFile{Dataset: ds, p: p}, nil
}

type imageFile struct {
	gdal.Dataset
	p *modis.ImageParams
}

func (ds *imageFile) ImageParams() *modis.ImageParams {
	return ds.p
}

func (ds *imageFile) Read(x, y int) (float64, error) {
	nx := ds.ImageParams().XSize()
	ny := ds.ImageParams().YSize()
	if x < 0 || x >= nx || y < 0 || y >= ny {
		return math.NaN(), fmt.Errorf("{x:%d, y:%d} is outside of image area {x:[0,%d), y:[,%d)}", x, y, nx, ny)
	}
	if res, err := ds.ReadBlock(x, y, modis.Box{0, 0, 1, 1}); err == nil {
		return res[0], nil
	} else {
		return math.NaN(), err
	}
}

func (ds *imageFile) ReadAtLatLon(ll modis.LatLon) (float64, error) {
	x, y := ds.ImageParams().Transform().LatLon2Pixels(ll)
	return ds.Read(x, y)
}

func (ds *imageFile) ReadTime(x, y int) (time.Time, error) {
	v, err := ds.Read(x, y)
	if err != nil {
		return time.Time{}, err
	}
	ll := ds.ImageParams().Transform().Pixels2LatLon(x, y)
	return ds.ImageParams().Value2time(v, ll)
}

func (ds *imageFile) ReadTimeAtLatLon(ll modis.LatLon) (time.Time, error) {
	v, err := ds.ReadAtLatLon(ll)
	if err != nil {
		return time.Time{}, err
	}
	return ds.ImageParams().Value2time(v, ll)
}

func (ds *imageFile) ReadBlock(x, y int, box modis.Box) ([]float64, error) {
	rb := ds.Dataset.RasterBand(band) // Assume 1 band or panic
	buffer := make([]float64, box[2]*box[3])
	err := rb.IO(gdal.Read, x+box[0], y+box[1], box[2], box[3], buffer, box[2], box[3], 0, 0)
	if err != nil {
		return nil, err
	}
	nan, hasnan := ds.ImageParams().NaN()
	for i, val := range buffer {
		if hasnan && buffer[i] == nan {
			buffer[i] = math.NaN()
		} else {
			buffer[i] = val*ds.ImageParams().Scale() + ds.ImageParams().Offset()
		}
	}
	return buffer, nil
}

func (ds *imageFile) ToMemory() *inMemory {
	res := NewInMemory(ds.ImageParams().ToBuilder().Build())
	// FIXME - copy data
	return res
}

func (ds *imageFile) Write(x, y int, v float64) error {
	return ds.WriteBlock(x, y, modis.Box{0, 0, 1, 1}, []float64{v})
}

func (ds *imageFile) WriteAtLatLon(ll modis.LatLon, v float64) error {
	x, y := ds.ImageParams().Transform().LatLon2Pixels(ll)
	return ds.Write(x, y, v)
}

func (ds *imageFile) WriteBlock(x, y int, box modis.Box, buffer []float64) error {
	rb := ds.Dataset.RasterBand(band) // Assume 1 band or panic
	// GDAL can handle any format, but it is more efficient to use specific type as we need to make a copy anyway
	switch ds.ImageParams().DataType() {
	case gdal.Int32:
		data := make([]int32, len(buffer))
		for i, v := range buffer {
			data[i] = int32((v - ds.ImageParams().Offset()) / ds.ImageParams().Scale())
		}
		return rb.IO(gdal.Write, x+box[0], y+box[1], box[2], box[3], data, box[2], box[3], 0, 0)
	case gdal.Float32:
		data := make([]float32, len(buffer))
		for i, v := range buffer {
			data[i] = float32((v - ds.ImageParams().Offset()) / ds.ImageParams().Scale())
		}
		return rb.IO(gdal.Write, x+box[0], y+box[1], box[2], box[3], data, box[2], box[3], 0, 0)
	default: // treat as float64
		data := make([]float64, len(buffer))
		for i, v := range buffer {
			data[i] = (v - ds.ImageParams().Offset()) / ds.ImageParams().Scale()
		}
		return rb.IO(gdal.Write, x+box[0], y+box[1], box[2], box[3], data, box[2], box[3], 0, 0)
	}
}

func (ds *imageFile) Close() {
	ds.Dataset.Close()
	ds.p = nil
}
