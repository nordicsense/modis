package ts

import (
	"regexp"
	"time"

	"github.com/nordicsense/gdal"
)

const (
	hdfPattern = `^.+\.hdf$`
)

var (
	sdsPattern = regexp.MustCompile(`^SUBDATASET_\d{1,}_NAME=(.+)$`)
)

type TimedDataset struct {
	Time    time.Time
	Dataset string
}

type LayerPair struct {
	Time  string
	Value string
}

type patternMatcher struct {
	timeMatcher  *regexp.Regexp
	valueMatcher *regexp.Regexp
}

// ListAll lists all pairs (time/value) of datasets contained within HDF files under
// root (recursive sub-folders) that match provided dataset name patterns. Only complete
// pairs are returned (incomplete or fully missing do not trigger error).
func ListAll(root string, layerPairPatterns ...LayerPair) ([]LayerPair, error) {
	var matchers []patternMatcher
	for _, layerPairPattern := range layerPairPatterns {
		timeMatcher, err := regexp.Compile(layerPairPattern.Time)
		if err != nil {
			return nil, err
		}
		valueMatcher, err := regexp.Compile(layerPairPattern.Value)
		if err != nil {
			return nil, err
		}
		matchers = append(matchers, patternMatcher{timeMatcher: timeMatcher, valueMatcher: valueMatcher})
	}

	hdfDSNames, err := ScanTree(root, hdfPattern)
	if err != nil {
		return nil, err
	}

	var layerPairs []LayerPair
	for _, hdfDSName := range hdfDSNames {
		pairs, err := getPairs(hdfDSName, matchers)
		if err != nil {
			return layerPairs, err
		}
		layerPairs = append(layerPairs, pairs...)
	}

	return layerPairs, err
}

func getPairs(hdfDSName string, matchers []patternMatcher) ([]LayerPair, error) {
	var layerPairs []LayerPair
	subHdfDSNames, err := listDatasets(hdfDSName)
	if err != nil {
		return nil, err
	}
	for _, matcher := range matchers {
		var timeDs string
		var valueDs string
		for _, subHDfDSName := range subHdfDSNames {
			if timeDs != "" && valueDs != "" {
				break
			}
			if matcher.timeMatcher.MatchString(subHDfDSName) {
				timeDs = subHDfDSName
				continue
			}
			if matcher.valueMatcher.MatchString(subHDfDSName) {
				valueDs = subHDfDSName
				continue
			}
		}
		if timeDs != "" && valueDs != "" {
			layerPairs = append(layerPairs, LayerPair{Time: timeDs, Value: valueDs})
		}
	}
	return layerPairs, nil
}

// listDatasets lists sub-datasets of a dataset (in particularly useful for HDF).
func listDatasets(dsname string) ([]string, error) {
	ds, err := gdal.Open(dsname, gdal.ReadOnly)
	if err != nil {
		return nil, err
	}
	defer ds.Close()

	var res []string
	for _, sds := range ds.Metadata("SUBDATASETS") {
		if sdsPattern.MatchString(sds) {
			res = append(res, sdsPattern.FindStringSubmatch(sds)[1])
		}
	}
	return res, nil
}
