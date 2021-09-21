package main

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Renderer struct {
	tasks       chan renderTask
	concurrency uint64
	closed      bool
}

type renderTask struct {
	ctx            context.Context
	pdfData        PDFGenerationData
	workdirPrepare func(context.Context) (*url.URL, error)
	result         chan interface{}
}

var (
	renderer *Renderer
)

func NewRenderer(concurrency uint64) *Renderer {
	br := &Renderer{
		tasks:       make(chan renderTask),
		concurrency: concurrency,
	}
	zap.L().Debug("renderer concurrency", zap.Uint64("concurrency", concurrency))
	go br.processTasks()
	return br
}

func (br *Renderer) Schedule(ctx context.Context, pdfData PDFGenerationData, prepareWorkdir func(context.Context) (*url.URL, error)) (buf []byte, err error) {
	resultCh := make(chan interface{})
	defer close(resultCh)
	br.tasks <- renderTask{ctx, pdfData, prepareWorkdir, resultCh}

	select {
	case <-ctx.Done():
		err = ctx.Err()
		return
	case res := <-resultCh:
		if rerr, ok := res.(error); ok {
			err = rerr
			return
		}
		buf = res.([]byte)
	}

	return
}

func (br *Renderer) Close() (err error) {
	if br.closed {
		return
	}
	br.closed = true
	close(br.tasks)
	return
}

func (br *Renderer) processTasks() {
	guard := make(chan struct{}, br.concurrency)
	for task := range br.tasks {
		if br.concurrency > 0 {
			guard <- struct{}{}
		}

		go func(task renderTask) {
			defer func() {
				if br.concurrency > 0 {
					<-guard
				}
			}()

			url, err := task.workdirPrepare(task.ctx)
			if err != nil {
				task.result <- fmt.Errorf("failed to prepare workdir: %w", err)
				return
			}

			buf, err := br.runChromeDP(task.ctx, url.String(), task.pdfData)
			if err != nil {
				task.result <- err
			} else {
				task.result <- buf
			}
		}(task)
	}
	zap.L().Debug("tasks channel closed")
}

func (br *Renderer) runChromeDP(ctx context.Context, url string, pdfData PDFGenerationData) (buf []byte, err error) {
	allocatorOpts := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)

	if noChromeSandbox {
		allocatorOpts = append(allocatorOpts, chromedp.NoSandbox)
	}

	// Set up logger
	logger := zap.NewStdLog(zap.L().With(zap.String("section", "chromedp"), zap.String("output", "stdout")))
	opts := []chromedp.ContextOption{
		chromedp.WithLogf(logger.Printf),
	}

	if debugMode {
		errLogger, _ := zap.NewStdLogAt(zap.L().With(zap.String("section", "chromedp"), zap.String("output", "stderr")), zapcore.WarnLevel)
		opts = append(opts, chromedp.WithErrorf(errLogger.Printf))

		// NOTE: very verbose
		/*
			debugLogger, _ := zap.NewStdLogAt(zap.L().With(zap.String("section", "chromedp"), zap.String("output", "debug")), zapcore.DebugLevel)
			opts = append(opts, chromedp.WithDebugf(debugLogger.Printf))
		*/
	}

	// Create separate timeout context
	timeoutCtx, tcancel := context.WithTimeout(ctx, time.Duration(rendererTimeout)*time.Millisecond)
	defer tcancel()

	// Create browser allocator
	allocator, cancel := chromedp.NewExecAllocator(timeoutCtx, allocatorOpts...)
	defer cancel()

	// Create context
	ctx, cancel2 := chromedp.NewContext(allocator, opts...)
	defer cancel2()

	// Capture pdf
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
			printParams := page.PrintToPDF().
				WithPrintBackground(true).
				WithLandscape(data.Layout.Orientation == OrientationLandscape).
				WithPreferCSSPageSize(true).
				WithDisplayHeaderFooter(false). // NOTE: keep this off at all times.
				WithPaperHeight(toInches(data.Layout.Size.Height)).
				WithPaperWidth(toInches(data.Layout.Size.Width)).
				WithMarginLeft(toInches(data.Layout.Margin.Left)).
				WithMarginTop(toInches(data.Layout.Margin.Top)).
				WithMarginRight(toInches(data.Layout.Margin.Right)).
				WithMarginBottom(toInches(data.Layout.Margin.Bottom))

			buf, _, err := printParams.Do(ctx)
			if err != nil {
				return err
			}

			*res = buf
			return nil
		}),
	}
}

func toInches(mm float64) float64 {
	return mm / 25.4
}
