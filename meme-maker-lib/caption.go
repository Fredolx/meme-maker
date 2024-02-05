package caption

import (
	"fmt"
	"path/filepath"
	"strings"

	imagick "gopkg.in/gographics/imagick.v3/imagick"
)

func fileNameWithoutExtension(fileName string) string {
	return strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
}

func AddCaption(filePath string, caption string, paddingPercent float64, font string, fontSize float64, output string) error {
	if output == "" {
		output = fmt.Sprintf("%s_memed%s", fileNameWithoutExtension(filePath), filepath.Ext(filePath))
	}
	if fontSize == 0 {
		fontSize = 72
	}
	imagick.Initialize()
	defer imagick.Terminate()
	var baseHeight float64 = fontSize*0.2 + fontSize
	mw := imagick.NewMagickWand()
	dw := imagick.NewDrawingWand()
	pw := imagick.NewPixelWand()
	var e = mw.ReadImage(filePath)
	if e != nil {
		return e
	}
	if font != "" {
		if e := dw.SetFont(font); e != nil {
			return e
		}
	} else {
		if e := dw.SetFontFamily("Courier"); e != nil {
			return e
		}
	}
	pw.SetColor("black")
	dw.SetTextAntialias(true)
	dw.SetFillColor(pw)
	dw.SetFontSize(fontSize)
	lines := splitCaption(caption, mw, dw, paddingPercent)
	mw.SetGravity(imagick.GRAVITY_NORTH)
	for i, str := range lines {
		y := float64(i * int(fontSize))
		dw.Annotation(0, y, str)
	}
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
		mw.SpliceImage(0, uint(baseHeight*float64(len(lines))), 0, 0)
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
		println(element)
	}
}
