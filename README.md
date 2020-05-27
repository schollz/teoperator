# teoperator


This repo is a [Go library](https://pkg.go.dev/github.com/schollz/teoperator/src/op1?tab=doc) and a server that you can chop up sounds that can build synth and drum patches for the OP-1 or OP-Z. I went down a [rabbit hole to reverse-engineer the OP-1 drum patch](https://schollz.com/blog/op1/) and this was the end result. You can access the server at https://op1z.com and you can build your own synth and drum patches from any sound.


<p align="center">
<a href="https://op1.schollz.com/patch?audioURL=https%3A%2F%2Fcdn.loc.gov%2Fservice%2Fgdc%2Fgdcarpl%2Fgdcarpl-1624415%2F1624415.mp3&secondsStart=982&secondsEnd=1002"><img src="/static/image/example2.png"></a>
</p>


## Install

You can follow these instructions to download this repo and run it locally. First install the pre-requisite programs. Instructions for both Windows and Linux.

#### Linux

```
$ sudo apt install imagemagick ffmpeg 
$ sudo add-apt-repository ppa:chris-needham/ppa
$ sudo apt-get update
$ sudo apt-get install audiowaveform
$ sudo -H python3 -m pip install youtube-dl
```

#### Windows

```
$ go get github.com/schollz/zget
$ zget https://github.com/wincentbalin/compile-static-audiowaveform/releases/download/1.2.2/audiowaveform-mingw64.zip
$ unzip audiowaveform-mingw64.zip 
$ mv audiowaveform to path
$ scoop install ffmpeg imagemagick youtube-dl
```

### Build and run

```
$ go build 
$ ./teoperator --serve --debug
[info]	2020/05/17 13:33:58 listening on :8053
```

Then open a browser to `localhost:8053`!


# License

MIT license

Please note THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
