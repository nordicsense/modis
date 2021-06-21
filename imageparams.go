package modis

import (
	"math"
	"time"

	"github.com/nordicsense/gdal"
)

type ImageParams struct {
	xSize      int
	ySize      int
	transform  AffineTransform
	projection string
	nan        float64
	nanPresent bool
	offset     float64
	scale      float64
	datatype   gdal.DataType
	metadata   map[string]string
	date       time.Time
}

func (ip *ImageParams) copy() *ImageParams {
	res := &ImageParams{
		xSize:      ip.xSize,
		ySize:      ip.ySize,
		transform:  ip.transform,
		projection: ip.projection,
		nan:        ip.nan,
		nanPresent: ip.nanPresent,
		offset:     ip.offset,
		scale:      ip.scale,
		datatype:   ip.datatype,
		metadata:   make(map[string]string),
		date:       ip.date,
	}
	for k, v := range ip.metadata {
		res.metadata[k] = v
	}
	return res
}

func (ip *ImageParams) ToBuilder() *imageParamsBuilder {
	return &imageParamsBuilder{ImageParams: ip.copy()}
}

func (ip *ImageParams) XSize() int {
	return ip.xSize
}

func (ip *ImageParams) YSize() int {
	return ip.ySize
}

func (ip *ImageParams) Transform() AffineTransform {
	return ip.transform
}

func (ip *ImageParams) Projection() string {
	return ip.projection
}

func (ip *ImageParams) NaN() (float64, bool) {
	return ip.nan, ip.nanPresent
}

func (ip *ImageParams) Offset() float64 {
	return ip.offset
}

func (ip *ImageParams) Scale() float64 {
	return ip.scale
}

func (ip *ImageParams) DataType() gdal.DataType {
	return ip.datatype
}

func (ip *ImageParams) Date() time.Time {
	return ip.date
}

func (ip *ImageParams) Metadata() map[string]string {
	return ip.metadata // TODO: copy or protect
}

func (ip *ImageParams) NorthWest() LatLon {
	return ip.Transform().Pixels2LatLon(0, 0)
}

func (ip *ImageParams) SouthEast() LatLon {
	return ip.Transform().Pixels2LatLon(ip.XSize()-1, ip.YSize()-1)
}

func (ip *ImageParams) Within(ll LatLon) bool {
	x, y := ip.Transform().LatLon2Pixels(ll)
	return x >= 0 && y >= 0 && x < ip.XSize() && y < ip.YSize()
}

func (ip *ImageParams) Value2time(v float64, ll LatLon) (time.Time, error) {
	utc := ModisLST2UTC(v, ll[1])
	return ip.Date().Add(time.Duration(utc * float64(time.Hour))), nil
}

func ImageParamsBuilder(xSize, ySize int) *imageParamsBuilder {
	ip := &ImageParams{
		xSize:      xSize,
		ySize:      ySize,
		transform:  AffineTransform{0, 1, 0, 0, 0, 1},
		projection: ModisWKT,
		offset:     0.0,
		scale:      1.0,
		nan:        math.NaN(),
		nanPresent: false,
		datatype:   gdal.Float64,
		metadata:   make(map[string]string),
		date:       time.Time{},
	}
	return &imageParamsBuilder{ImageParams: ip}
}

type imageParamsBuilder struct {
	*ImageParams
}

func (ipb *imageParamsBuilder) Transform(transform AffineTransform) *imageParamsBuilder {
	ipb.transform = transform
	return ipb
}

func (ipb *imageParamsBuilder) Projection(projection string) *imageParamsBuilder {
	ipb.projection = projection
	return ipb
}

func (ipb *imageParamsBuilder) NaN(nan float64) *imageParamsBuilder {
	ipb.nanPresent = true
	ipb.nan = nan
	return ipb
}

func (ipb *imageParamsBuilder) Offset(offset float64) *imageParamsBuilder {
	ipb.offset = offset
	return ipb
}

func (ipb *imageParamsBuilder) Scale(scale float64) *imageParamsBuilder {
	ipb.scale = scale
	return ipb
}

func (ipb *imageParamsBuilder) DataType(dt gdal.DataType) *imageParamsBuilder {
	ipb.datatype = dt
	return ipb
}

func (ipb *imageParamsBuilder) Date(tm time.Time) *imageParamsBuilder {
	ipb.date = tm
	return ipb
}

func (ipb *imageParamsBuilder) Metadata(key, value string) *imageParamsBuilder {
	ipb.metadata[key] = value
	return ipb
}

func (ipb *imageParamsBuilder) Build() *ImageParams {
	return ipb.ImageParams.copy()
}
