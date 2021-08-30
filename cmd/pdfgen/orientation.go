package main

import (
	"encoding/json"
	"fmt"
)

type Orientation string

func (o *Orientation) UnmarshalJSON(data []byte) (err error) {
	var value string
	if err = json.Unmarshal(data, &value); err != nil {
		return
	}
	*o = Orientation(value)
	if _, ok := Orientations[*o]; !ok {
		err = fmt.Errorf("invalid orientation '%s'", value)
	}
	return
}

var (
	OrientationLandscape Orientation = "landscape"
	OrientationPortrait  Orientation = "portrait"
)

var Orientations = map[Orientation]bool{
	OrientationLandscape: true,
	OrientationPortrait:  true,
}
