# teoperator

I reverse-engineered [the OP-1 drum patch](https://github.com/schollz/teoperator/blob/master/src/op1/op1.go#L52-L129) so you can build your own drum patches from the OP-1. This repo is a server that you can chop up sounds from the internet for easy loading into the OP-1. Try it at https://op1.schollz.com.

## Install

You can also install and run yourself:

```
$ sudo apt install imagemagick ffmpeg 
$ sudo add-apt-repository ppa:chris-needham/ppa
$ sudo apt-get update
$ sudo apt-get install audiowaveform
$ sudo -H python3 -m pip install youtube-dl
```

#### Windows

```
go get github.com/schollz/zget
zget https://github.com/wincentbalin/compile-static-audiowaveform/releases/download/1.2.2/audiowaveform-mingw64.zip
unzip audiowaveform-mingw64.zip 
mv audiowaveform to path
scoop install ffmpeg imagemagick youtube-dl
```

# License

MIT license

Please note THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.