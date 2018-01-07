package uda

import (
	"fmt"
	"github.com/alpacahq/marketstore/utils/io"
)

func ColumnToFloat32(cols io.ColumnInterface, name string) (outCol []float32, err error) {
	ccol := cols.GetColumn(name)
	if ccol == nil {
		return nil, fmt.Errorf("Unable to retrieve column named %s", name)
	}
	switch cc := ccol.(type) {
	case []float32:
		outCol = cc
	case []float64:
		outCol = make([]float32, len(cc))
		for i := range cc {
			outCol[i] = float32(cc[i])
		}
	case []int:
		outCol = make([]float32, len(cc))
		for i := range cc {
			outCol[i] = float32(cc[i])
		}
	case []int64:
		outCol = make([]float32, len(cc))
		for i := range cc {
			outCol[i] = float32(cc[i])
		}
	case []int32:
		outCol = make([]float32, len(cc))
		for i := range cc {
			outCol[i] = float32(cc[i])
		}
	}
	return outCol, nil
}
