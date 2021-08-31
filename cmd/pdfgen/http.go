package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/valyala/fasthttp"
)

func wrap(handler func(ctx *fasthttp.RequestCtx) error) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		if err := handler(ctx); err != nil {
			ctx.Error(err.Error(), http.StatusInternalServerError)
		}
	}
}

func unmarshalJson(data []byte, target interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(&target)
}
