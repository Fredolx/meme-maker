package caption

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	imagick "gopkg.in/gographics/imagick.v3/imagick"
)

type CaptionArgs struct {
	FilePath        string
	Caption         string
	BottomCaption   string
	XPaddingPercent float64
	YPaddingPercent float64
	Font            string
	FontSize        float64
	LineHeightPx    float64
	Color           string
	StrokeWidth     float64
	Output          string
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
	if e := wands.mw.ReadImage(caption.FilePath); e != nil {
		return e
	}
	isGif := wands.mw.GetImageFormat() == "GIF"
	setDefaults(&caption, float64(wands.mw.GetImageWidth()))
	setUpDrawingWands(&caption, &wands)
	if caption.YPaddingPercent != 0 {
		yPaddingPx = caption.FontSize * (caption.YPaddingPercent / 100)
	}
	if wands.bottomDW != nil {
		bottomCaptionLines := splitCaption(caption.Caption, caption.XPaddingPercent,
			wands.mw, wands.bottomDW)
		_ = annotate(AnnotateArgs{
			mw:           wands.mw,
			dw:           wands.bottomDW,
			lines:        bottomCaptionLines,
			yPaddingPx:   yPaddingPx,
			gravity:      imagick.GRAVITY_SOUTH,
			lineHeightPx: caption.LineHeightPx,
		})
	}
	lines := splitCaption(caption.Caption, caption.XPaddingPercent, wands.mw, wands.dw)
	topCaptionHeight := annotate(AnnotateArgs{
		mw:           wands.mw,
		dw:           wands.dw,
		lines:        lines,
		yPaddingPx:   yPaddingPx,
		gravity:      imagick.GRAVITY_NORTH,
		lineHeightPx: caption.LineHeightPx,
	})
	if isGif {
		handleGifs(wands.mw)
	}
	if e := drawImage(wands, topCaptionHeight); e != nil {
		return e
	}
	if e := writeImage(isGif, wands, caption.Output); e != nil {
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

func handleGifs(mw *imagick.MagickWand) {
	mw = mw.CoalesceImages()
	var bgColor, _ = mw.GetImageBackgroundColor()
	if bgColor.GetAlpha() == 0 {
		mw.SetImageDispose(imagick.DISPOSE_BACKGROUND)
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
	if caption.Color == "" {
		caption.Color = "black"
	}
	if caption.Output == "" {
		caption.Output = fmt.Sprintf("%s_memed%s", fileNameWithoutExtension(caption.FilePath),
			filepath.Ext(caption.FilePath))
	}
	if caption.Font == "" {
		if runtime.GOOS == "windows" {
			caption.Font = "Arial"
		} else {
			caption.Font = "DejaVu-Sans"
		}
	}
	if caption.FontSize == 0 {
		caption.FontSize = float64(imageWidth) / 15
	}
	if caption.StrokeWidth == 0 {
		caption.StrokeWidth = caption.FontSize / 40
	}
}

func setUpDrawingWands(caption *CaptionArgs, wands *CaptionWands) error {
	var topBottomMode = caption.BottomCaption != ""
	var e error
	if topBottomMode {
		caption.Color = "white"
		if wands.bottomDW, e = setUpDrawingWand(caption, wands.pw); e != nil {
			return e
		}
	}
	if wands.dw, e = setUpDrawingWand(caption, wands.pw); e != nil {
		return e
	}
	return nil
}

func setUpDrawingWand(caption *CaptionArgs, pw *imagick.PixelWand) (*imagick.DrawingWand, error) {
	dw := imagick.NewDrawingWand()
	if e := dw.SetFont(caption.Font); e != nil {
		return nil, e
	}
	pw.SetColor(caption.Color)
	dw.SetFillColor(pw)
	if caption.BottomCaption != "" {
		pw.SetColor("black")
		dw.SetStrokeColor(pw)
		dw.SetStrokeWidth(caption.StrokeWidth)
		dw.SetStrokeAntialias(true)
	}
	dw.SetFontSize(caption.FontSize)
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
