package dataset_test

import (
	"path"
	"testing"

	"github.com/nordicsense/modis/lib/progress"
	"github.com/nordicsense/modis/dataset"
	"github.com/lukeroth/gdal"

	modists "github.com/nordicsense/modis/ts"
)

func testinit(t *testing.T) []modists.LayerPair {
	dataRoot := "/data/Research/Data"
	aquaRoot := path.Join(dataRoot, "MODIS", "MOD11A1", "Aqua-HDF")
	pairs, err := modists.ListAll(aquaRoot, modists.LayerPair{
		Time:  `^.+_LST:Day_view_time$`,
		Value: `^.+_LST:LST_Day_.+$`,
	})
	if err != nil {
		t.Fatal(err)
	}
	return pairs
}

func TestGDALOpenClose_MemoryLeak(t *testing.T) {
	pairs := testinit(t)
	bar := progress.Start("test", len(pairs))
	for _, pair := range pairs {
		ds, err := gdal.Open(pair.Value, gdal.ReadOnly)
		if err != nil {
			t.Fatal(err)
		}
		rb := ds.RasterBand(1)
		for j := 0; j < 300; j++ {
			buffer := make([]float64, 1)
			if err = rb.IO(gdal.Read, j+100, j+100, 1, 1, buffer, 1, 1, 0, 0); err != nil {
				t.Fatal(err)
			}
		}
		ds.Close()
		bar.Add(1)
	}
}

func TestOpenClose_MemoryLeak(t *testing.T) {
	pairs := testinit(t)
	bar := progress.Start("test", len(pairs))
	for _, pair := range pairs {
		ds, err := dataset.Open(pair.Time)
		if err != nil {
			t.Fatal(err)
		}
		for j := 0; j < 300; j++ {
			if _, err := ds.ReadTime(j+100, j+100); err != nil {
				t.Fatal(err)
			}
		}
		ds.Close()
		bar.Add(1)
	}
}
