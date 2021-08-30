package main

import (
	"encoding/json"
	"fmt"
)

// Based on https://pspdfkit.com/guides/processor/pdf-generation/pdf-generation-schema/
/*
type Orientation = "landscape" | "portrait";
type PageSize =
  | "A0"
  | "A1"
  | "A2"
  | "A3"
  | "A4"
  | "A5"
  | "A6"
  | "A7"
  | "A8"
  | "Letter"
  | "Legal";

type PdfGenerationSchema = {
  html: string, // The HTML file passed in the multipart request.
  assets?: Array<string>, // All assets imported in the HTML. Reference the name passed in the multipart request.
  layout?: {
    orientation?: Orientation,
    size?: {
      width: number,
      height: number
    } | PageSize, // {width, height} in mm or page size preset.
    margin?: {
      // Margin sizes in mm.
      left: number,
      top: number,
      right: number,
      bottom: number
    }
  }
};
*/

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
	Layout struct {
		Orientation Orientation `json:"orientation"`
		Size        struct {
			Width  float64   `json:"width"`
			Height float64   `json:"height"`
			Preset *PageSize `json:"preset"`
		} `json:"size"`
		Margin struct {
			Left   float64 `json:"left"`
			Top    float64 `json:"top"`
			Right  float64 `json:"right"`
			Bottom float64 `json:"bottom"`
		} `json:"margin"`
	} `json:"layout"`
}

func NewPDFGenerationData(payload PDFGenerationSchema) (d PDFGenerationData, err error) {
	pageSize := PageSizeA4

	d.Html = payload.Html
	d.Assets = append([]string{}, payload.Assets...)

	d.Layout.Orientation = OrientationPortrait
	d.Layout.Size.Preset, d.Layout.Size.Width, d.Layout.Size.Height = pageSize.Preset(d.Layout.Orientation)

	d.Layout.Margin.Left = 10
	d.Layout.Margin.Top = 10
	d.Layout.Margin.Right = 10
	d.Layout.Margin.Bottom = 10

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
				d.Layout.Size.Preset, d.Layout.Size.Width, d.Layout.Size.Height = ps.Preset(d.Layout.Orientation)
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
