package main

import (
	"encoding/json"
	"fmt"
)

type PDFGenerationSchema struct {
	// The HTML file passed in the multipart request.
	Html string `json:"html"`
	// All assets imported in the HTML. Reference the name passed in the multipart request.
	Assets []string `json:"assets"`
	Layout *struct {
		Orientation *Orientation `json:"orientation"`
		// { width: mm, height: mm } or anything from PageSize constants
		Size   *json.RawMessage `json:"size"`
		Margin *struct {
			Left   float64 `json:"left"`
			Top    float64 `json:"top"`
			Right  float64 `json:"right"`
			Bottom float64 `json:"bottom"`
		} `json:"margin"`
	} `json:"layout"`
}

// Normalized version of `PDFGenerationSchema`
type PDFGenerationData struct {
	Html   string   `json:"html"`
	Assets []string `json:"assets"`
	Layout layout   `json:"layout"`
}

type layout struct {
	Orientation Orientation `json:"orientation"`
	Size        PDFSize     `json:"size"`
	Margin      PDFMargin   `json:"margin"`
}

type PDFSize struct {
	Width  float64   `json:"width"`
	Height float64   `json:"height"`
	Preset *PageSize `json:"preset"`
}

type PDFMargin struct {
	Left   float64 `json:"left"`
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
}

func NewPDFGenerationData(payload PDFGenerationSchema) (d PDFGenerationData, err error) {
	pageSize := PageSizeA4

	d.Html = payload.Html
	d.Assets = append([]string{}, payload.Assets...)

	d.Layout.Orientation = OrientationPortrait
	d.Layout.Size.Preset, d.Layout.Size.Width, d.Layout.Size.Height = pageSize.Preset()

	d.Layout.Margin.Left = 0
	d.Layout.Margin.Top = 0
	d.Layout.Margin.Right = 0
	d.Layout.Margin.Bottom = 0

	if l := payload.Layout; l != nil {
		// Try to copy orientation
		if o := l.Orientation; o != nil {
			if _, ok := Orientations[*o]; ok {
				d.Layout.Orientation = Orientation(*o)
			} else {
				err = fmt.Errorf("unknown orientation '%s'", *o)
				return
			}
		}

		// Unmarshal the size
		if s := l.Size; s != nil {
			s := *s
			if len(s) > 0 && s[0] == '{' { // TODO: hacky
				var size struct {
					Width  float64 `json:"width"`
					Height float64 `json:"height"`
				}
				if err = json.Unmarshal(s, &size); err != nil {
					err = fmt.Errorf("failed to unmarshal page size: %w", err)
					return
				}
				d.Layout.Size.Width, d.Layout.Size.Height = size.Width, size.Height
			} else {
				var ps PageSize
				if err = ps.UnmarshalJSON(s); err != nil {
					err = fmt.Errorf("failed to unmarshal page size: %w", err)
					return
				}
				d.Layout.Size.Preset, d.Layout.Size.Width, d.Layout.Size.Height = ps.Preset()
			}
		}

		// Unmarshal the margins info
		if m := l.Margin; m != nil {
			d.Layout.Margin.Left = m.Left
			d.Layout.Margin.Top = m.Top
			d.Layout.Margin.Right = m.Right
			d.Layout.Margin.Bottom = m.Bottom

		}
	}

	return
}
