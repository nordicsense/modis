package ts_test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/nordicsense/modis/ts"
)

func TestListAll(t *testing.T) {
	const nKeys = 15
	const dayKey = "MODIS_Grid_Daily_1km_LST:Day_view_time MODIS_Grid_Daily_1km_LST:LST_Day_1km"
	const nightKey = "MODIS_Grid_Daily_1km_LST:Night_view_time MODIS_Grid_Daily_1km_LST:LST_Night_1km"

	dir, _ := os.Getwd()
	dsPairs, err := ts.ListAll(path.Join(dir, "..", "..", "testdata"), ts.LSTDay, ts.LSTNight)
	if err != nil {
		t.Fatal(err)
	}
	res := make(map[string]int)
	for _, dsPair := range dsPairs {
		p := strings.SplitAfter(dsPair.Time, ".hdf\":")[1] + " " + strings.SplitAfter(dsPair.Value, ".hdf\":")[1]
		res[p]++
	}
	if len(res) != 2 {
		t.Errorf("len(keys)=%d, want 2", len(res))
	}
	if res[dayKey] != nKeys {
		t.Errorf("# day keys = %d, want %d", res[dayKey], nKeys)
	}
	if res[nightKey] != nKeys {
		t.Errorf("# night keys = %d, want %d", res[nightKey], nKeys)
	}
}

func TestReadDatasetTimeseries(t *testing.T) {
	from := time.Date(2013, 8, 19, 0, 0, 0, 0, time.UTC)
	to := time.Date(2016, 2, 3, 0, 0, 0, 0, time.UTC)
	opts := ts.Options{
		LayerPatterns: []ts.LayerPair{ts.LSTNight},
		FromDate:      &from,
		ToDate:        &to,
	}
	dir, _ := os.Getwd()
	dsTs, err := ts.ReadDatasetTimeseries(path.Join(dir, "..", "..", "testdata"), opts)
	if err != nil {
		t.Fatal(err)
	}
	for i, ds := range dsTs {
		dsTs[i].Time = dsTs[i].Time.Round(time.Minute)
		dsTs[i].Dataset = strings.SplitAfter(ds.Dataset, ".hdf\":")[1]
	}
	const want = "[{`2013-08-19 19:01:00 +0000 UTC` `MODIS_Grid_Daily_1km_LST:LST_Night_1km`}" +
		" {`2013-08-20 18:06:00 +0000 UTC` `MODIS_Grid_Daily_1km_LST:LST_Night_1km`}" +
		" {`2013-08-21 20:28:00 +0000 UTC` `MODIS_Grid_Daily_1km_LST:LST_Night_1km`}" +
		" {`2016-02-01 19:01:00 +0000 UTC` `MODIS_Grid_Daily_1km_LST:LST_Night_1km`}" +
		" {`2016-02-02 19:47:00 +0000 UTC` `MODIS_Grid_Daily_1km_LST:LST_Night_1km`}" +
		" {`2016-02-03 18:51:00 +0000 UTC` `MODIS_Grid_Daily_1km_LST:LST_Night_1km`}]"
	if fmt.Sprintf("%#q", dsTs) != want {
		t.Errorf("ReadDatasetTimeseries()=%#q, want=%s", dsTs, want)
	}

}
