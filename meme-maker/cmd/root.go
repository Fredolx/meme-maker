package cmd

import (
	"log"

	caption "github.com/fredolx/meme-maker-lib"
	"github.com/spf13/cobra"
)

const (
	fontConst               = "font"
	fontShortConst          = "f"
	fontSizeConst           = "size"
	fontSizeShortConst      = "s"
	xPaddingConst           = "pad-x"
	xPaddingShortConst      = "x"
	yPaddingConst           = "pad-y"
	yPaddingShortConst      = "y"
	lineHeightConst         = "line"
	lineHeightShortConst    = "l"
	outputConst             = "output"
	outputShortConst        = "o"
	bottomCaptionConst      = "bottom-caption"
	bottomCaptionShortConst = "b"
)

var rootCmd = &cobra.Command{
	Use:     "meme-maker",
	Short:   "Add captions to memes",
	Args:    cobra.MinimumNArgs(2),
	Example: `meme-maker myimage.png "my caption"`,
	Run: func(cmd *cobra.Command, args []string) {
		var xPadding, _ = cmd.Flags().GetFloat64(xPaddingConst)
		var yPadding, _ = cmd.Flags().GetFloat64(yPaddingConst)
		var font, _ = cmd.Flags().GetString(fontConst)
		var fontSize, _ = cmd.Flags().GetFloat64(fontSizeConst)
		var lineHeight, _ = cmd.Flags().GetFloat64(lineHeightConst)
		var bottomCaption, _ = cmd.Flags().GetString(bottomCaptionConst)
		var output, _ = cmd.Flags().GetString(outputConst)
		if e := caption.AddCaption(args[0], args[1], bottomCaption, xPadding, yPadding, font, fontSize, lineHeight, output); e != nil {
			log.Fatal(e)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringP(fontConst, fontShortConst, "", "Font to use")
	rootCmd.Flags().Float64P(fontSizeConst, fontSizeShortConst, 0, "Font size")
	rootCmd.Flags().Float64P(xPaddingConst, xPaddingShortConst, 5, "Horizontal padding percentage")
	rootCmd.Flags().Float64P(yPaddingConst, yPaddingShortConst, 30, "Vertical padding percentage")
	rootCmd.Flags().Float64P(lineHeightConst, lineHeightShortConst, 0, "Line height in pixels")
	rootCmd.Flags().StringP(bottomCaptionConst, bottomCaptionShortConst, "", "Bottom caption")
	rootCmd.Flags().StringP(outputConst, outputShortConst, "", "Output path for the meme'd image")
}
