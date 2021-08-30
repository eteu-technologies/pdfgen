package main

import (
	"encoding/json"
	"fmt"
)

type PageSize string

func (p *PageSize) UnmarshalJSON(data []byte) (err error) {
	var value string
	if err = json.Unmarshal(data, &value); err != nil {
		return
	}
	*p = PageSize(value)
	if _, ok := PageSizes[*p]; !ok {
		err = fmt.Errorf("invalid page size preset '%s'", value)
	}
	return
}

func (p *PageSize) Preset(or Orientation) (*PageSize, float64, float64) {
	return PageSizes[*p](or)
}

type PageSizePreset func(or Orientation) (*PageSize, float64, float64)

func newPreset(name PageSize, w, h float64) PageSizePreset {
	return func(or Orientation) (*PageSize, float64, float64) {
		switch or {
		case OrientationPortrait:
			return &name, w, h
		case OrientationLandscape:
			return &name, h, w
		default:
			panic("unreachable")
		}
	}
}

var (
	PageSizeA0     PageSize = "A0"
	PageSizeA1     PageSize = "A1"
	PageSizeA2     PageSize = "A2"
	PageSizeA3     PageSize = "A3"
	PageSizeA4     PageSize = "A4"
	PageSizeA5     PageSize = "A5"
	PageSizeA6     PageSize = "A6"
	PageSizeA7     PageSize = "A7"
	PageSizeA8     PageSize = "A8"
	PageSizeLetter PageSize = "Letter"
	PageSizeLegal  PageSize = "Legal"
)

var PageSizes = map[PageSize]PageSizePreset{
	PageSizeA0:     newPreset(PageSizeA0, 841, 1189),
	PageSizeA1:     newPreset(PageSizeA1, 594, 841),
	PageSizeA2:     newPreset(PageSizeA2, 420, 594),
	PageSizeA3:     newPreset(PageSizeA3, 297, 420),
	PageSizeA4:     newPreset(PageSizeA4, 210, 297),
	PageSizeA5:     newPreset(PageSizeA5, 148.5, 210),
	PageSizeA6:     newPreset(PageSizeA6, 105, 148.5),
	PageSizeA7:     newPreset(PageSizeA7, 74, 105),
	PageSizeA8:     newPreset(PageSizeA8, 52, 74),
	PageSizeLetter: newPreset(PageSizeLetter, 215.9, 279.4), // 8.5" x 11"
	PageSizeLegal:  newPreset(PageSizeLegal, 215.9, 335.6),  // 8.5" x 14"
}
