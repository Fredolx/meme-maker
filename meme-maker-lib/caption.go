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

type CaptionWands struct {
	mw       *imagick.MagickWand
	pw       *imagick.PixelWand
	dw       *imagick.DrawingWand
	bottomDW *imagick.DrawingWand
}

type AnnotateArgs struct {
	mw           *imagick.MagickWand
	dw           *imagick.DrawingWand
	lines        []string
	yPaddingPx   float64
	gravity      imagick.GravityType
	lineHeightPx float64
}

func fileNameWithoutExtension(fileName string) string {
	return strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
}

func AddCaption(caption CaptionArgs) error {
	var yPaddingPx float64 = 0
	var wands CaptionWands = CaptionWands{}
	imagick.Initialize()
	defer imagick.Terminate()
	wands.mw = imagick.NewMagickWand()
	wands.pw = imagick.NewPixelWand()
	if e := wands.mw.ReadImage(caption.filePath); e != nil {
		return e
	}
	isGif := wands.mw.GetImageFormat() == "GIF"
	setDefaults(&caption, float64(wands.mw.GetImageWidth()))
	setUpDrawingWands(&caption, wands)
	if caption.yPaddingPercent != 0 {
		yPaddingPx = caption.fontSize * (caption.yPaddingPercent / 100)
	}
	if wands.bottomDW != nil {
		bottomCaptionLines := splitCaption(caption.caption, caption.xPaddingPercent,
			wands.mw, wands.bottomDW)
		_ = annotate(AnnotateArgs{
			mw:           wands.mw,
			dw:           wands.bottomDW,
			lines:        bottomCaptionLines,
			yPaddingPx:   yPaddingPx,
			gravity:      imagick.GRAVITY_SOUTH,
			lineHeightPx: caption.lineHeightPx,
		})
	}
	lines := splitCaption(caption.caption, caption.xPaddingPercent, wands.mw, wands.dw)
	topCaptionHeight := annotate(AnnotateArgs{
		mw:           wands.mw,
		dw:           wands.dw,
		lines:        lines,
		yPaddingPx:   yPaddingPx,
		gravity:      imagick.GRAVITY_NORTH,
		lineHeightPx: caption.lineHeightPx,
	})
	if isGif {
		handleGifs(wands)
	}
	drawImage(wands, topCaptionHeight)
	if e := writeImage(isGif, wands, caption.output); e != nil {
		return e
	}
	return nil
}

func writeImage(isGif bool, wands CaptionWands, output string) error {
	if isGif {
		if e := wands.mw.WriteImages(output, true); e != nil {
			return e
		}
	} else {
		if e := wands.mw.WriteImage(output); e != nil {
			return e
		}
	}
	return nil
}

func handleGifs(wands CaptionWands) {
	wands.mw = wands.mw.CoalesceImages()
	var bgColor, _ = wands.mw.GetImageBackgroundColor()
	if bgColor.GetAlpha() == 0 {
		wands.mw.SetImageDispose(imagick.DISPOSE_BACKGROUND)
	}
}

func drawImage(wands CaptionWands, topCaptionHeight float64) error {
	for ok := true; ok; ok = wands.mw.NextImage() {
		wands.pw.SetColor("white")
		wands.mw.SetImageBackgroundColor(wands.pw)
		if wands.bottomDW == nil {
			wands.mw.SpliceImage(0, uint(topCaptionHeight), 0, 0)
		} else {
			if e := wands.mw.DrawImage(wands.bottomDW); e != nil {
				return e
			}
		}
		if e := wands.mw.DrawImage(wands.dw); e != nil {
			return e
		}
	}
	return nil
}

func setDefaults(caption *CaptionArgs, imageWidth float64) {
	if caption.output == "" {
		caption.output = fmt.Sprintf("%s_memed%s", fileNameWithoutExtension(caption.filePath),
			filepath.Ext(caption.filePath))
	}
	if caption.font == "" {
		if runtime.GOOS == "windows" {
			caption.font = "Arial"
		} else {
			caption.font = "DejaVu-Sans"
		}
	}
	if caption.fontSize == 0 {
		caption.fontSize = float64(imageWidth) / 15
	}
	if caption.strokeWidth == 0 {
		caption.strokeWidth = caption.fontSize / 40
	}
}

func setUpDrawingWands(caption *CaptionArgs, wands CaptionWands) error {
	var topBottomMode = caption.bottomCaption != ""
	var e error
	if topBottomMode {
		caption.color = "white"
		if wands.bottomDW, e = setUpDrawingWand(caption, wands); e != nil {
			return e
		}
	}
	if wands.dw, e = setUpDrawingWand(caption, wands); e != nil {
		return e
	}
	return nil
}

func setUpDrawingWand(caption *CaptionArgs, wands CaptionWands) (*imagick.DrawingWand, error) {
	dw := imagick.NewDrawingWand()
	if e := dw.SetFont(caption.font); e != nil {
		return nil, e
	}
	wands.pw.SetColor(caption.color)
	dw.SetFillColor(wands.pw)
	wands.pw.SetColor("black")
	dw.SetStrokeColor(wands.pw)
	dw.SetStrokeWidth(caption.strokeWidth)
	dw.SetStrokeAntialias(true)
	dw.SetFontSize(caption.fontSize)
	return dw, nil
}

func annotate(args AnnotateArgs) float64 {
	args.dw.SetGravity(args.gravity)
	metrics := args.mw.QueryFontMetrics(args.dw, args.lines[0])
	var y float64 = metrics.BoundingBoxY2 - metrics.Ascender + args.yPaddingPx
	if args.gravity == imagick.GRAVITY_SOUTH {
		y = args.yPaddingPx
	}
	for i, str := range args.lines {
		if i != 0 {
			y += args.mw.QueryFontMetrics(args.dw, str).TextHeight + args.lineHeightPx
		}
		args.dw.Annotation(0, y, str)
	}
	if strings.ContainsAny(args.lines[len(args.lines)-1], "gjpqy") {
		y += metrics.TextHeight
	} else {
		y += metrics.Ascender
	}
	y += args.yPaddingPx
	return y
}

func splitCaption(caption string, xPaddingPercent float64, mw *imagick.MagickWand, dw *imagick.DrawingWand) []string {
	var currentWidth float64 = 0
	lines := []string{}
	width := float64(mw.GetImageWidth())
	if xPaddingPercent != 0 {
		width = width - width*(xPaddingPercent/100)
	}
	spaceWidth := mw.QueryFontMetrics(dw, " ").TextWidth
	words := strings.Split(caption, " ")
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
