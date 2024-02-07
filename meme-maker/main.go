package main

import (
	"log"
	"meme-maker/cmd"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	log.SetFlags(0)
	if runtime.GOOS == "windows" {
		exePath, _ := os.Executable()
		base := os.Getenv("MEMEMAKER_BASE_PATH")
		if base == "" {
			base = filepath.Join(filepath.Dir(exePath), "..", "lib", "ImageMagick-7.1.1", "modules-Q16HDRI")
		}
		coders := filepath.Join(base, "coders")
		filters := filepath.Join(base, "filters")
		os.Setenv("MAGICK_CODER_MODULE_PATH", coders)
		os.Setenv("MAGICK_CODER_FILTER_PATH", filters)
	}
	cmd.Execute()
}
