package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strconv"
	"sync"

	"github.com/AubSs/fasthttplogger"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/davecgh/go-spew/spew"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

var (
	listenAddr = ":5000"
	workdirs   sync.Map
)

func main() {
	sigCh := make(chan os.Signal, 1)
	exitCh := make(chan bool, 1)
	signal.Notify(sigCh, os.Interrupt)

	router := router.New()
	srv := fasthttp.Server{
		Handler: fasthttplogger.CombinedColored(router.Handler),
	}

	router.POST("/process", wrap(HandleProcess))
	router.GET("/serve/{key}/{path:*}", wrap(HandleServe))

	go func() {
		if err := srv.ListenAndServe(listenAddr); err != nil {
			log.Println("failed to listen for http:", err)
			exitCh <- true
		}
	}()

	select {
	case <-sigCh:
		log.Println("got signal")
	case <-exitCh:
		// no-op
	}

	if err := srv.Shutdown(); err != nil {
		log.Println("failed to shutdown the http server:", err)
	}
}

func HandleProcess(ctx *fasthttp.RequestCtx) (err error) {
	pdfName := "processed.pdf"

	// Parse pdf data
	var pdfSchema PDFGenerationSchema
	if err = json.Unmarshal(ctx.FormValue("generation"), &pdfSchema); err != nil {
		return
	}

	pdfData, err := NewPDFGenerationData(pdfSchema)
	if err != nil {
		return err
	}

	var targetURL *url.URL
	if rawURL := ctx.FormValue("url"); len(rawURL) > 0 {
		targetURL, err = url.Parse(string(rawURL))
		if err != nil {
			return err
		}

		pdfName = targetURL.Hostname() + ".pdf"
	} else {
		workdir, mainFile, err, cleanupFn := prepareWorkdir(ctx, pdfData)
		if err != nil {
			return err
		}

		// TODO: nicer way to get workdir key
		workdirKey := path.Base(workdir)
		workdirs.Store(workdirKey, workdir)

		log.Printf("created workdir '%s'", workdirKey)

		defer func() {
			log.Printf("cleaning up workdir '%s'", workdirKey)
			workdirs.Delete(workdirKey)
			cleanupFn()
		}()
		targetURL, _ = url.Parse(fmt.Sprintf(`http://127.0.0.1:5000/serve/%s/%s`, workdirKey, mainFile))
	}

	pdfBytes, err := runChromeDP(ctx, targetURL.String(), pdfData)
	if err != nil {
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

	// Strip two slashes, /serve/{key}/...
	workdir := rawWorkdir.(string)
	log.Printf("wd=%s", workdir)
	fasthttp.FSHandler(workdir, 2)(ctx)
	return
}

func runChromeDP(ctx context.Context, url string, pdfData PDFGenerationData) (buf []byte, err error) {
	// create context
	ctx, cancel := chromedp.NewContext(ctx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// capture pdf
	if err = chromedp.Run(ctx, printToPDF(url, pdfData, &buf)); err != nil {
		return
	}

	return
}

// print a specific pdf page.
func printToPDF(url string, data PDFGenerationData, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {

			spew.Dump(data)

			printParams := page.PrintToPDF().
				WithPrintBackground(false).
				WithLandscape(data.Layout.Orientation == OrientationLandscape).
				WithPreferCSSPageSize(true).
				WithDisplayHeaderFooter(false). // NOTE: keep this off at all times.
				WithPaperHeight(float64(data.Layout.Size.Height) / 25.4).
				WithPaperWidth(float64(data.Layout.Size.Width) / 25.4).
				WithMarginLeft(float64(data.Layout.Margin.Left) / 25.4).
				WithMarginTop(float64(data.Layout.Margin.Top) / 25.4).
				WithMarginRight(float64(data.Layout.Margin.Right) / 25.4).
				WithMarginBottom(float64(data.Layout.Margin.Bottom) / 25.4)

			buf, _, err := printParams.Do(ctx)
			if err != nil {
				return err
			}

			*res = buf
			return nil
		}),
	}
}
