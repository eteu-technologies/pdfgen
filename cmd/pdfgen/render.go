package main

import (
	"context"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func runChromeDP(ctx context.Context, url string, pdfData PDFGenerationData) (buf []byte, err error) {
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
