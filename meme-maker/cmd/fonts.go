package cmd

import (
	caption "github.com/fredolx/meme-maker-lib"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(fonts)
}

var fonts = &cobra.Command{
	Use:     "fonts",
	Aliases: []string{"f"},
	Args:    cobra.ArbitraryArgs,
	Short:   "Lists and filters all fonts present on the system",
	Long:    "Lists and filters all fonts present on the system.\nIf the application hangs while using a font, it means this font is incompatible.",
	Run: func(cmd *cobra.Command, args []string) {
		var font = ""
		if len(args) > 0 {
			font = args[0]
		}
		caption.ListFonts(font)
	},
}
