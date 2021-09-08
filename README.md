# pdfgen

[![Lint](https://github.com/eteu-technologies/pdfgen/actions/workflows/lint.yml/badge.svg)](https://github.com/eteu-technologies/pdfgen/actions/workflows/lint.yml)
[![Build](https://github.com/eteu-technologies/pdfgen/actions/workflows/build.yml/badge.svg)](https://github.com/eteu-technologies/pdfgen/actions/workflows/build.yml)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Feteu-technologies%2Fpdfgen.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Feteu-technologies%2Fpdfgen?ref=badge_shield)

[![built with nix](https://builtwithnix.org/badge.svg)](https://builtwithnix.org)

Microservice for generating PDF files from HTML and assets easily.

This project is using Chromium to render, so you are free to use everything what Chromium supports (HTML5 and CSS Flex for example).

pdfgen implements a subset of PSPDFKit Processor's API and features (no support for existing PDF files and operations, only
generation from scratch; no special handling for inputs (no fillable forms)).

## Running

We're publishing x86_64 Linux Docker images on [Docker Hub](https://hub.docker.com/r/eteu/pdfgen).

aarch64 Linux images are planned.

Recommended way to run the generator microservice:

```bash
docker run --rm -ti \
    --name pdfgen \
    --read-only \
    --tmpfs /tmp \
    --shm-size 2gb \
    -p 5000:5000 \
    eteu/pdfgen:latest
```

If possible, also consider dropping internet access from the container.

## API documentation

`POST /process` with multipart/form-data

Based on https://pspdfkit.com/guides/processor/pdf-generation/pdf-generation-schema/

```typescript
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
```

### curl example

Requires [jq](https://stedolan.github.io/jq/) to be installed as well.

```bash
#!/usr/bin/env bash
set -euo pipefail

gen='{
    "html": "invoice.html"
,   "assets": [
       "bootstrap-5.0.2.min.css"
    ]
,   "layout": {
        "orientation": "portrait"
,       "size": "A4"
,       "margin": {
            "top": 10
        ,   "bottom": 10
        ,   "left": 10
        ,   "right": 10
        }
    }
}'

curl -v \
    -X POST \
    -o invoice.pdf \
    -F generation="$(jq -c <<< "${gen}")" \
    -F bootstrap-5.0.2.min.css=@./bootstrap-5.0.2.min.css \
    -F invoice.html=@./invoice.html \
	http://127.0.0.1:5000/process
```

## License

LGPL 3.0


[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Feteu-technologies%2Fpdfgen.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Feteu-technologies%2Fpdfgen?ref=badge_large)
