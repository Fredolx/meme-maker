package caption

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	imagick "gopkg.in/gographics/imagick.v3/imagick"
)

type CaptionArgs struct {
	filePath        string
	caption         string
	bottomCaption   string
	xPaddingPercent float64
	yPaddingPercent float64
	font            string
	fontSize        float64
	lineHeightPx    float64
	color           string
	strokeWidth     float64
	output          string
}

func fileNameWithoutExtension(fileName string) string {
	return strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
}

func AddCaption(caption CaptionArgs) error {
	var topBottomMode = bottomCaption != ""
	var color = "black"
	var e error
	var dw, bottomDW *imagick.DrawingWand
	var yPaddingPx float64 = 0
	var isGif bool
	if output == "" {
		output = fmt.Sprintf("%s_memed%s", fileNameWithoutExtension(filePath), filepath.Ext(filePath))
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
	pw := imagick.NewPixelWand()
	if e = mw.ReadImage(filePath); e != nil {
		return e
	}
	if fontSize == 0 {
		fontSize = float64(mw.GetImageWidth()) / 15
	}
	isGif = mw.GetImageFormat() == "GIF"
	if yPaddingPercent != 0 {
		yPaddingPx = fontSize * (yPaddingPercent / 100)
	}
	if topBottomMode {
		color = "white"
		if bottomDW, e = setUpDrawingWand(pw, font, fontSize, color); e != nil {
			return e
		}
		bottomCaptionLines := splitCaption(bottomCaption, mw, bottomDW, xPaddingPercent)
		_ = annotate(mw, bottomDW, bottomCaptionLines, yPaddingPx, lineHeightPx, imagick.GRAVITY_SOUTH)
	}
	if dw, e = setUpDrawingWand(pw, font, fontSize, color); e != nil {
		return e
	}
	lines := splitCaption(caption, mw, dw, xPaddingPercent)
	topCaptionHeight := annotate(mw, dw, lines, yPaddingPx, lineHeightPx, imagick.GRAVITY_NORTH)
	if isGif {
		mw = mw.CoalesceImages()
		var bgColor, _ = mw.GetImageBackgroundColor()
		if bgColor.GetAlpha() == 0 {
			mw.SetImageDispose(imagick.DISPOSE_BACKGROUND)
		}
	}
	for ok := true; ok; ok = mw.NextImage() {
		pw.SetColor("white")
		mw.SetImageBackgroundColor(pw)
		if !topBottomMode {
			mw.SpliceImage(0, uint(topCaptionHeight), 0, 0)
		} else {
			if e := mw.DrawImage(bottomDW); e != nil {
				return e
			}
		}
		if e = mw.DrawImage(dw); e != nil {
			return e
		}
	}
	if isGif {
		if e = mw.WriteImages(output, true); e != nil {
			return e
		}
	} else {
		if e = mw.WriteImage(output); e != nil {
			return e
		}
	}
	return nil
}

func setUpDrawingWand(pw *imagick.PixelWand, font string, fontSize float64, color string, stroke bool) (*imagick.DrawingWand, error) {
	dw := imagick.NewDrawingWand()
	if e := dw.SetFont(font); e != nil {
		return nil, e
	}
	pw.SetColor(color)
	dw.SetFillColor(pw)
	pw.SetColor("black")
	dw.SetStrokeColor(pw)
	dw.SetStrokeWidth(fontSize / 40)
	dw.SetStrokeAntialias(true)
	dw.SetFontSize(fontSize)
	return dw, nil
}

func annotate(mw *imagick.MagickWand, dw *imagick.DrawingWand, lines []string, yPaddingPx float64, lineHeightPx float64, gravity imagick.GravityType) float64 {
	dw.SetGravity(gravity)
	metrics := mw.QueryFontMetrics(dw, lines[0])
	var y float64 = metrics.BoundingBoxY2 - metrics.Ascender + yPaddingPx
	if gravity == imagick.GRAVITY_SOUTH {
		y = yPaddingPx
	}
	for i, str := range lines {
		if i != 0 {
			y += mw.QueryFontMetrics(dw, str).TextHeight + lineHeightPx
		}
		dw.Annotation(0, y, str)
	}
	if strings.ContainsAny(lines[len(lines)-1], "gjpqy") {
		y += metrics.TextHeight
	} else {
		y += metrics.Ascender
	}
	y += yPaddingPx
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
