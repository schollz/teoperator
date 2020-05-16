# teoperator

I reverse-engineered [the OP-1 drum patch](https://github.com/schollz/teoperator/blob/master/src/op1/op1.go#L52-L129) so you can build your own drum patches from the OP-1. This repo is a server that you can chop up sounds from the internet for easy loading into the OP-1. The server is currently hosted here: https://op1.schollz.com

<strong>Examples:</strong> <a href="https://op1.schollz.com/patch?audioURL=https%3A%2F%2Fcdn.loc.gov%2Fservice%2Fgdc%2Fgdcarpl%2Fgdcarpl-1624415%2F1624415.mp3&secondsStart=982&secondsEnd=1002">poetry from library of congress</a>, <a href="https://op1.schollz.com/patch?audioURL=https%3A%2F%2Fupload.wikimedia.org%2Fwikipedia%2Fcommons%2F6%2F68%2FTurdus_merula_male_song_at_dawn%252820s%2529.ogg&secondsStart=0&secondsEnd=30">black bird sounds</a>, <a href="https://op1.schollz.com/patch?audioURL=https%3A%2F%2Fupload.wikimedia.org%2Fwikipedia%2Fcommons%2F7%2F70%2FTimpani_64-c-p5.wav&secondsStart=0&secondsEnd=0">sounds from a timpani</a>. <a href="https://op1.schollz.com/patch?audioURL=https%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3D36CYMdFmDeQ&secondsStart=21.9&secondsEnd=60">spoken word from youtube.com</a>.

<p align="center">
<a href="https://op1.schollz.com/patch?audioURL=https%3A%2F%2Fcdn.loc.gov%2Fservice%2Fgdc%2Fgdcarpl%2Fgdcarpl-1624415%2F1624415.mp3&secondsStart=982&secondsEnd=1002"><img src="/static/image/example.png"></a>
</p>


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
