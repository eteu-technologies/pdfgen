package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

func prepareWorkdir(ctx *fasthttp.RequestCtx, data PDFGenerationData) (workdir string, mainFile string, err error, cleanupFn func()) {
	form, err := ctx.MultipartForm()
	if err != nil {
		return
	}

	workdir, err = ioutil.TempDir("", "pdfgenwd")
	if err != nil {
		return
	}

	cleanupFn = func() {
		if workdir != "" {
			if err := os.RemoveAll(workdir); err != nil {
				zap.L().Error("failed to clean up workdir", zap.String("workdir", workdir), zap.Error(err))
			}
		}
	}

	files := map[string][]byte{}

	// Collect files
	mainFile = path.Base(data.Html)
	if mainFile == "" {
		err = fmt.Errorf("generation html is empty")
		return
	}

	if files[mainFile], err = getFile(form, data.Html); err != nil {
		go cleanupFn()
		return
	}

	for _, name := range data.Assets {
		name := path.Base(name)

		if _, ok := files[name]; ok {
			err = fmt.Errorf("duplicate file '%s'", name)
			go cleanupFn()
			return
		}

		if files[name], err = getFile(form, name); err != nil {
			go cleanupFn()
			return
		}
	}

	// Copy files
	for name, bytes := range files {
		target := path.Join(workdir, name)
		if err = ioutil.WriteFile(target, bytes, 0600); err != nil {
			err = fmt.Errorf("failed to write '%s': %w", target, err)
			go cleanupFn()
			return
		}
	}

	return
}
