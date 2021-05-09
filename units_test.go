package modis_test

import (
	"math"
	"testing"

	"github.com/arcticgeo/modis"
)

func TestLatLon_Transform(t *testing.T) {
	ll := modis.LatLon{67.97, 32.9}
	expected := modis.LatLon{7.557919164030658e+06, 1.372204024867538e+06}
	actual, err := ll.Transform(4326, 53008)
	assertLatLon(t, expected, actual, err)
}

func TestLatLon_Degrees2Sin(t *testing.T) {
	ll := modis.LatLon{67.97, 32.9}
	expected := modis.LatLon{7.557919164030658e+06, 1.372204024867538e+06}
	actual, err := ll.Degrees2Sin()
	assertLatLon(t, expected, actual, err)
}

func TestLatLon_Sin2Degree(t *testing.T) {
	expected := modis.LatLon{67.97, 32.9}
	ll := modis.LatLon{7.557919164030658e+06, 1.372204024867538e+06}
	actual, err := ll.Sin2Degree()
	assertLatLon(t, expected, actual, err)
}

func assertLatLon(t *testing.T, expected modis.LatLon, actual modis.LatLon, err error) {
	if err != nil {
		t.Error(err)
		return
	}
	if math.Abs(expected[0]-actual[0]) > 1e-5 {
		t.Errorf("Expected lat %v, found %v", expected[0], actual[0])
	}
	if math.Abs(expected[1]-actual[1]) > 1e-5 {
		t.Errorf("Expected lon %v, found %v", expected[1], actual[1])
	}
}

func TestModisLST2UTC(t *testing.T) {
	cases := []struct {
		lst       float64
		lonDegree float64
		expected  float64
	}{
		{
			lst:       -1.25,
			lonDegree: 33.2,
			expected:  20.536666666,
		},
		{
			lst:       9.45,
			lonDegree: 33.9,
			expected:  7.1899999999,
		},
		{
			lst:       25.3,
			lonDegree: 45.1,
			expected:  -1.7066666666,
		},
	}
	for _, data := range cases {
		actual := modis.ModisLST2UTC(data.lst, data.lonDegree)
		if math.Abs(data.expected-actual) > 1e-5 {
			t.Errorf("Expected UTC %v, found %v for LST %v", data.expected, actual, data.lst)
		}
	}
}

func TestAffineTransform_LatLon2Pixels(t *testing.T) {
	// 2002-07-04
	tf := modis.AffineTransform{1111950, 926, 0.0, 7783653, 0.0, -926}
	ll, _ := modis.LatLon{60.25, 25.05}.Degrees2Sin()
	x, y := tf.LatLonSin2Pixels(ll)
	if x != 292 || y != 1171 {
		t.Errorf("expected (292,1171), found (%d, %d)", x, y)
	}
	x, y = tf.LatLon2Pixels(modis.LatLon{60.25, 25.05})
	if x != 292 || y != 1171 {
		t.Errorf("expected (292,1171), found (%d, %d)", x, y)
	}
}
