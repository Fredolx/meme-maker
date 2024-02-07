package cmd

import (
	"log"

	caption "github.com/fredolx/meme-maker-lib"
	"github.com/spf13/cobra"
)

const (
	fontConst            = "font"
	fontShortConst       = "f"
	fontSizeConst        = "size"
	fontSizeShortConst   = "s"
	xPaddingConst        = "pad-x"
	xPaddingShortConst   = "x"
	yPaddingConst        = "pad-y"
	yPaddingShortConst   = "y"
	lineHeightConst      = "line"
	lineHeightShortConst = "l"
	outputConst          = "output"
	outputShortConst     = "o"
)

var rootCmd = &cobra.Command{
	Use:     "meme-maker",
	Short:   "Add captions to memes",
	Args:    cobra.ExactArgs(2),
	Example: `meme-maker myimage.png "my caption"`,
	Run: func(cmd *cobra.Command, args []string) {
		var xPadding, _ = cmd.Flags().GetFloat64(xPaddingConst)
		var yPadding, _ = cmd.Flags().GetFloat64(yPaddingConst)
		var font, _ = cmd.Flags().GetString(fontConst)
		var fontSize, _ = cmd.Flags().GetFloat64(fontSizeConst)
		var lineHeight, _ = cmd.Flags().GetFloat64(lineHeightConst)
		var output, _ = cmd.Flags().GetString(outputConst)
		if e := caption.AddCaption(args[0], args[1], xPadding, yPadding, font, fontSize, lineHeight, output); e != nil {
			log.Fatal(e)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringP(fontConst, fontShortConst, "", "Sets the font to use for the caption")
	rootCmd.Flags().Float64P(fontSizeConst, fontSizeShortConst, 50, "Sets the font size for the caption")
	rootCmd.Flags().Float64P(xPaddingConst, xPaddingShortConst, 5, "Sets the horizontal padding percentage for the caption")
	rootCmd.Flags().Float64P(yPaddingConst, yPaddingShortConst, 30, "Sets the vertical padding percentage for the caption")
	rootCmd.Flags().Float64P(lineHeightConst, lineHeightShortConst, 0, "Sets the line height in pixels for the caption")
	rootCmd.Flags().StringP(outputConst, outputShortConst, "", "Sets the output path for the meme'd image")
}
