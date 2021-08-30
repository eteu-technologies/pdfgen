package main

import (
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
