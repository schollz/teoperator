experiments:

- can i manipulate the op-1 meta data directly and still upload (try on 1.aif)?
- can convert 1.aif into a new aif with ffmpeg (11.aif) and inject an OP-1 payload into it before the SSND?

## Install

```
$ sudo apt install imagemagick ffmpeg 
$ sudo add-apt-repository ppa:chris-needham/ppa
$ sudo apt-get update
$ sudo apt-get install audiowaveform
```

#### Windows

```
go get github.com/schollz/zget
zget https://github.com/wincentbalin/compile-static-audiowaveform/releases/download/1.2.2/audiowaveform-mingw64.zip
unzip audiowaveform-mingw64.zip 
mv audiowaveform to path
scoop install ffmpeg imagemagick
```

```
ffmpeg -i 1.aif -af silencedetect=noise=-30dB:d=0.1 -f null -
```