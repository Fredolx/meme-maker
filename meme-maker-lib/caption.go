package caption

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	imagick "gopkg.in/gographics/imagick.v3/imagick"
)

func fileNameWithoutExtension(fileName string) string {
	return strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
}

func AddCaption(filePath string, caption string, xPaddingPercent float64, yPaddingPercent float64, font string,
	fontSize float64, lineHeightPx float64, output string) error {
	if output == "" {
		output = fmt.Sprintf("%s_memed%s", fileNameWithoutExtension(filePath), filepath.Ext(filePath))
	}
	if fontSize == 0 {
		fontSize = 72
	}
	if font == "" {
		if runtime.GOOS == "windows" {
			font = "Arial"
		} else {
			font = "DejaVu-Sans"
		}
	}
	imagick.Initialize()
	defer imagick.Terminate()
	mw := imagick.NewMagickWand()
	dw := imagick.NewDrawingWand()
	pw := imagick.NewPixelWand()
	var e = mw.ReadImage(filePath)
	if e != nil {
		return e
	}
	if e := dw.SetFont(font); e != nil {
		return e
	}
	pw.SetColor("black")
	dw.SetFillColor(pw)
	dw.SetFontSize(fontSize)
	lines := splitCaption(caption, mw, dw, xPaddingPercent)
	var yPaddingPx float64 = 0
	if yPaddingPercent != 0 {
		yPaddingPx = fontSize * (yPaddingPercent / 100)
	}
	totalHeight := annotate(mw, dw, lines, yPaddingPx, lineHeightPx)
	if mw.GetImageFormat() == "GIF" {
		mw = mw.CoalesceImages()
		var bgColor, _ = mw.GetImageBackgroundColor()
		if bgColor.GetAlpha() == 0 {
			mw.SetImageDispose(imagick.DISPOSE_BACKGROUND)
		}
	}
	for ok := true; ok; ok = mw.NextImage() {
		pw.SetColor("white")
		mw.SetImageBackgroundColor(pw)
		mw.SpliceImage(0, uint(totalHeight), 0, 0)
		if e := mw.DrawImage(dw); e != nil {
			return e
		}
	}
	if mw.GetImageFormat() == "GIF" {
		if e := mw.WriteImages(output, true); e != nil {
			return e
		}
	} else {
		if e := mw.WriteImage(output); e != nil {
			return e
		}
	}
	return nil
}

func annotate(mw *imagick.MagickWand, dw *imagick.DrawingWand, lines []string, yPaddingPx float64, lineHeightPx float64) float64 {
	mw.SetGravity(imagick.GRAVITY_NORTH)
	metrics := mw.QueryFontMetrics(dw, lines[0])
	var y float64 = metrics.BoundingBoxY2 - metrics.Ascender + yPaddingPx
	fmt.Printf("%f %f %f %f, %f", metrics.BoundingBoxY2, metrics.TextHeight, metrics.CharacterHeight, metrics.Ascender, metrics.Descender)
	for i, str := range lines {
		if i != 0 {
			y += mw.QueryFontMetrics(dw, str).Ascender + lineHeightPx
		}
		dw.Annotation(0, y, str)
	}
	y += metrics.TextHeight + yPaddingPx
	return y
}

func splitCaption(caption string, mw *imagick.MagickWand, dw *imagick.DrawingWand, paddingPercent float64) []string {
	lines := []string{}
	width := float64(mw.GetImageWidth())
	if paddingPercent != 0 {
		width = width - width*(paddingPercent/100)
	}
	spaceWidth := mw.QueryFontMetrics(dw, " ").TextWidth
	words := strings.Split(caption, " ")
	var currentWidth float64 = 0
	currentLine := 0
	for _, word := range words {
		txtWidth := mw.QueryFontMetrics(dw, word).TextWidth
		if currentWidth == 0 {
			currentWidth += txtWidth
		} else {
			currentWidth += txtWidth + spaceWidth
		}
		if currentWidth < width {
			if len(lines)-1 != currentLine {
				lines = append(lines, word)
			} else {
				lines[currentLine] += " " + word
			}
		} else {
			currentLine += 1
			currentWidth = txtWidth
			lines = append(lines, word)
		}
	}
	return lines
}

func ListFonts(font string) {
	imagick.Initialize()
	defer imagick.Terminate()
	mw := imagick.NewMagickWand()
	font += "*"
	array := mw.QueryFonts(font)
	for _, element := range array {
		fmt.Println(element)
	}
}
