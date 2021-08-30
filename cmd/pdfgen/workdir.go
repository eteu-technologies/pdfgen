package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"os"
	"path"

	"github.com/valyala/fasthttp"
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
				log.Printf("failed to clean up workdir '%s': %s", workdir, err)
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
	return ioutil.ReadAll(file)
}
