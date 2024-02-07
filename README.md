# meme-maker
## Adding captions to images in one easy command

https://github.com/Fredolx/meme-maker/assets/38048855/d3a93ac6-e175-49bd-ac83-a5e09bcb16e2

## Installation
[Both Windows and Linux builds are available here](https://github.com/Fredolx/meme-maker/releases/latest).


The Windows build is batteries-included (has all dependencies pre-packaged), the linux build only depends on ImageMagick7.

If you want to see an official scoop version for Windows, star this repo! (50 are required for being on Extras). 

In the meantime, you can still use scoop to install meme-maker using this repo's manifest:
```
scoop install https://raw.githubusercontent.com/Fredolx/meme-maker/main/meme-maker.json
```

## Usage
meme-maker prides itself in its simplicity. You can generate a quick meme using only two arguments:
```
usage: meme-maker [path to file] [your meme caption]
meme-maker myimage.png "my caption"
```
More customizations can be achieved through optional flags which can be discovered through:
```
meme-maker -h
```
