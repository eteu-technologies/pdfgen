package main

import (
	"context"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/davecgh/go-spew/spew"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func runChromeDP(ctx context.Context, url string, pdfData PDFGenerationData) (buf []byte, err error) {
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

	// create context
	ctx, cancel := chromedp.NewContext(ctx, opts...)
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
