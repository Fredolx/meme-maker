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

type Caption struct {
	args          CaptionArgs
	wands         CaptionWands
	isGif         bool
	topBottomMode bool
	imageWidth    float64
	yPaddingPx    float64
}

func fileNameWithoutExtension(fileName string) string {
	return strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
}

func (c *Caption) AddCaption(args CaptionArgs) error {
	c.args = args
	c.wands = CaptionWands{}
	imagick.Initialize()
	defer imagick.Terminate()
	c.wands.mw = imagick.NewMagickWand()
	c.wands.pw = imagick.NewPixelWand()
	if e := c.wands.mw.ReadImage(c.args.FilePath); e != nil {
		return e
	}
	c.getImageProps()
	c.setDefaults()
	c.setUpDrawingWands()
	if c.topBottomMode {
		bottomCaptionLines := c.splitCaption(c.wands.bottomDW, c.args.BottomCaption)
		_ = c.annotate(c.wands.bottomDW, bottomCaptionLines, imagick.GRAVITY_SOUTH)
	}
	lines := c.splitCaption(c.wands.dw, c.args.Caption)
	topCaptionHeight := c.annotate(c.wands.dw, lines, imagick.GRAVITY_NORTH)
	if c.isGif {
		c.handleGifs()
	}
	if e := c.drawImage(topCaptionHeight); e != nil {
		return e
	}
	if e := c.writeImage(); e != nil {
		return e
	}
	return nil
}

func (c *Caption) getImageProps() {
	c.imageWidth = float64(c.wands.mw.GetImageWidth())
	c.isGif = c.wands.mw.GetImageFormat() == "GIF"
	c.topBottomMode = c.args.BottomCaption != ""
}

func (c *Caption) writeImage() error {
	if c.isGif {
		if e := c.wands.mw.WriteImages(c.args.Output, true); e != nil {
			return e
		}
	} else {
		if e := c.wands.mw.WriteImage(c.args.Output); e != nil {
			return e
		}
	}
	return nil
}

func (c *Caption) handleGifs() {
	c.wands.mw = c.wands.mw.CoalesceImages()
	var bgColor, _ = c.wands.mw.GetImageBackgroundColor()
	if bgColor.GetAlpha() == 0 {
		c.wands.mw.SetImageDispose(imagick.DISPOSE_BACKGROUND)
	}
}

func (c *Caption) drawImage(topCaptionHeight float64) error {
	for ok := true; ok; ok = c.wands.mw.NextImage() {
		if !c.topBottomMode {
			c.addWhiteSpace(topCaptionHeight)
		} else {
			if e := c.wands.mw.DrawImage(c.wands.bottomDW); e != nil {
				return e
			}
		}
		if e := c.wands.mw.DrawImage(c.wands.dw); e != nil {
			return e
		}
	}
	return nil
}

func (c *Caption) addWhiteSpace(topCaptionHeight float64) {
	c.wands.pw.SetColor("white")
	c.wands.mw.SetImageBackgroundColor(c.wands.pw)
	c.wands.mw.SpliceImage(0, uint(topCaptionHeight), 0, 0)
}

func (c *Caption) setDefaults() {
	if c.args.Color == "" {
		c.args.Color = "black"
	}
	if c.args.Output == "" {
		c.args.Output = fmt.Sprintf("%s_memed%s", fileNameWithoutExtension(c.args.FilePath),
			filepath.Ext(c.args.FilePath))
	}
	if c.args.Font == "" {
		if runtime.GOOS == "windows" {
			c.args.Font = "Arial"
		} else {
			c.args.Font = "DejaVu-Sans"
		}
	}
	if c.args.FontSize == 0 {
		c.args.FontSize = float64(c.imageWidth) / 15
	}
	if c.args.YPaddingPercent != 0 {
		c.yPaddingPx = c.args.FontSize * (c.args.YPaddingPercent / 100)
	}
	if c.args.StrokeWidth == 0 {
		c.args.StrokeWidth = c.args.FontSize / 35
	}
}

func (c *Caption) setUpDrawingWands() error {
	var e error
	if c.topBottomMode {
		c.args.Color = "white"
		if c.wands.bottomDW, e = c.setUpDrawingWand(); e != nil {
			return e
		}
	}
	if c.wands.dw, e = c.setUpDrawingWand(); e != nil {
		return e
	}
	return nil
}

func (c *Caption) setUpDrawingWand() (*imagick.DrawingWand, error) {
	dw := imagick.NewDrawingWand()
	if e := dw.SetFont(c.args.Font); e != nil {
		return nil, e
	}
	c.wands.pw.SetColor(c.args.Color)
	dw.SetFillColor(c.wands.pw)
	if c.topBottomMode {
		c.wands.pw.SetColor("black")
		dw.SetStrokeColor(c.wands.pw)
		dw.SetStrokeWidth(c.args.StrokeWidth)
		dw.SetStrokeAntialias(true)
	}
	dw.SetFontSize(c.args.FontSize)
	return dw, nil
}

func (c *Caption) annotate(dw *imagick.DrawingWand, lines []string, gravity imagick.GravityType) float64 {
	dw.SetGravity(gravity)
	metrics := c.wands.mw.QueryFontMetrics(dw, lines[0])
	var y float64 = metrics.BoundingBoxY2 - metrics.Ascender + c.yPaddingPx
	if gravity == imagick.GRAVITY_SOUTH {
		y = c.yPaddingPx
	}
	for i, str := range lines {
		if i != 0 {
			y += c.wands.mw.QueryFontMetrics(dw, str).TextHeight + c.args.LineHeightPx
		}
		dw.Annotation(0, y, str)
	}
	if strings.ContainsAny(lines[len(lines)-1], "gjpqy") {
		y += metrics.TextHeight
	} else {
		y += metrics.Ascender
	}
	y += c.yPaddingPx
	return y
}

func (c *Caption) splitCaption(dw *imagick.DrawingWand, caption string) []string {
	var currentWidth float64 = 0
	lines := []string{}
	width := c.imageWidth
	if c.args.XPaddingPercent != 0 {
		width = width - width*(c.args.XPaddingPercent/100)
	}
	spaceWidth := c.wands.mw.QueryFontMetrics(dw, " ").TextWidth
	words := strings.Split(caption, " ")
	currentLine := 0
	for _, word := range words {
		txtWidth := c.wands.mw.QueryFontMetrics(dw, word).TextWidth
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
