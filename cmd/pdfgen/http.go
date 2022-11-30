package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

func wrap(handler func(ctx *fasthttp.RequestCtx) error) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		defer func() {
			if p := recover(); p != nil {
				zap.L().With(zap.String("section", "http")).Error("recovered from panic", zap.Reflect("err", p))
				ctx.Error(fmt.Sprintf("panic: %+v", p), http.StatusInternalServerError)
			}
		}()
		if err := handler(ctx); err != nil {
			zap.L().With(zap.String("section", "http")).Error("handler error", zap.Error(err))
			ctx.Error(err.Error(), http.StatusInternalServerError)
		}
	}
}

func unmarshalJson(data []byte, target interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(&target)
}

func getFile(form *multipart.Form, name string) (data []byte, err error) {
	files := form.File[name]
	if len(files) == 0 {
		err = fmt.Errorf("file '%s' was not supplied", name)
		return
	}

	file, err := files[0].Open()
	if err != nil {
		err = fmt.Errorf("failed to read file '%s'", name)
	}

	defer func() { _ = file.Close() }()
	return io.ReadAll(file)
}

func requestLogger(req fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		begin := time.Now()
		req(ctx)
		end := time.Now()

		zap.L().With(zap.String("section", "http")).Debug(
			string(ctx.Method()),
			zap.String("address", ctx.RemoteAddr().String()),
			zap.Int("status", ctx.Response.Header.StatusCode()),
			zap.String("uri", string(ctx.RequestURI())),
			zap.String("useragent", string(ctx.UserAgent())),
			zap.Duration("duration", end.Sub(begin)),
		)
	})
}
