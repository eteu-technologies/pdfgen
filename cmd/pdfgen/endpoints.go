package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

func HandleProcess(ctx *fasthttp.RequestCtx) (err error) {
	pdfName := "processed.pdf"

	// Parse pdf data
	var pdfSchema PDFGenerationSchema
	if err = unmarshalJson(ctx.FormValue("generation"), &pdfSchema); err != nil {
		return
	}

	pdfData, err := NewPDFGenerationData(pdfSchema)
	if err != nil {
		return err
	}

	var workdirKey string
	var cleanupFn func()

	var targetURL *url.URL
	if rawURL := ctx.FormValue("url"); len(rawURL) > 0 {
		targetURL, err = url.Parse(string(rawURL))
		if err != nil {
			return err
		}

		pdfName = targetURL.Hostname() + ".pdf"
	}

	defer func() {
		if workdirKey != "" {
			zap.L().Debug("cleaning up", zap.String("key", workdirKey))
			workdirs.Delete(workdirKey)
			cleanupFn()
		}
	}()

	pdfBytes, err := renderer.Schedule(ctx, pdfData, func(wctx context.Context) (*url.URL, error) {
		if targetURL != nil {
			return targetURL, nil
		}

		var workdir, mainFile string
		workdir, mainFile, err, cleanupFn = prepareWorkdir(ctx, pdfData)
		if err != nil {
			return nil, err
		}

		// TODO: nicer way to get workdir key
		workdirKey = path.Base(workdir)
		workdirs.Store(workdirKey, workdir)

		zap.L().Debug("created workdir", zap.String("key", workdirKey))
		targetURL, _ = url.Parse(fmt.Sprintf(`http://%s/%s/%s`, iServerAddr, workdirKey, mainFile))

		return targetURL, nil
	})
	if errors.Is(err, context.DeadlineExceeded) {
		ctx.Error(err.Error(), http.StatusServiceUnavailable)
		return nil
	} else if err != nil {
		return err
	}

	ctx.SetStatusCode(http.StatusOK)
	ctx.SetContentType("application/pdf")
	ctx.Response.Header.Add("Content-Disposition", `inline; filename=`+strconv.Quote(pdfName))
	ctx.SetBody(pdfBytes)

	return
}

func HandleServe(ctx *fasthttp.RequestCtx) (err error) {
	workdirKey := ctx.UserValue("key")
	rawWorkdir, ok := workdirs.Load(workdirKey)
	if !ok {
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}

	// Strip one slash, /{key}/...
	workdir := rawWorkdir.(string)
	zap.L().Debug("serving workdir", zap.String("key", workdirKey.(string)), zap.String("path", workdir))
	fasthttp.FSHandler(workdir, 1)(ctx)
	return
}
