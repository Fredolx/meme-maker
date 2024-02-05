package cmd

import (
	"log"

	caption "github.com/fredolx/meme-maker-lib"
	"github.com/spf13/cobra"
)

const (
	fontConst          = "font"
	fontShortConst     = "f"
	fontSizeConst      = "size"
	fontSizeShortConst = "s"
	paddingConst       = "padding"
	paddingShortConst  = "p"
	outputConst        = "output"
	outputShortConst   = "o"
)

var rootCmd = &cobra.Command{
	Use:     "meme-maker",
	Short:   "Add captions to memes",
	Args:    cobra.ExactArgs(2),
	Example: `meme-maker myimage.png "my caption"`,
	Run: func(cmd *cobra.Command, args []string) {
		var padding, _ = cmd.Flags().GetFloat64(paddingShortConst)
		var font, _ = cmd.Flags().GetString(fontConst)
		var fontSize, _ = cmd.Flags().GetFloat64(fontSizeConst)
		var output, _ = cmd.Flags().GetString(outputConst)
		if e := caption.AddCaption(args[0], args[1], padding, font, fontSize, output); e != nil {
			log.Fatal(e)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringP(fontConst, fontShortConst, "", "Sets the font to use for the caption")
	rootCmd.Flags().Float64P(fontSizeConst, fontSizeShortConst, 72, "Sets the font size for the caption")
	rootCmd.Flags().Float64P(paddingConst, paddingShortConst, 5, "Sets the padding percentage for the caption")
	rootCmd.Flags().StringP(outputConst, outputShortConst, "", "Sets the output path for the meme'd image")
}
