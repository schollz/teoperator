# teoperator

*teoperator* lets you easily make drum and synth patches for the op-1 or op-z. You don't need this software to use *teoperator* - you can use most of the functionality via [teoperator.com](https://teoperator.com). You can also use this software via a command-line program that gives you more functionality - to create a variety of patches from any number of music files (wav, aif, mp3, flac all supported). I did a [write-up of how it works](https://schollz.com/blog/op1/), basically I had to reverse engineer pieces of the metadata in the op-1 and op-z patches. 

## Installation

*teoperator* requires ffmpeg. First, [install ffmpeg](https://ffmpeg.org/download.html). 

To use as a command line program you first need to [install Go](https://golang.org/doc/install) and then in a terminal run:

```
go get -v github.com/schollz/teoperator@latest
```

That will install `teoperator` on your system.

## Make synth patches

To make a synth patch just type:

```
teoperator synth piano.wav
```

Optionally, you can include the base frequency information which can be used on the op-1/opz to convert to the right pitch:

```
teoperator synth --freq 220 piano.wav
```

## Make drum patches

### Make a drum kit patch

To make a drumkit patch you can convert multiple files and splice points will be set at the boundaries of each individual file:

```
teoperator drum kick.wav snare.wav openhat.wav closedhat.wav
```

### Make a sample patch

To make a sample patch you can convert one sample and splice points will be automatically determined by transients:

```
teoperator drum vocals.wav
```

### Make a drum loop patch

To make a drum loop patch you can convert one sample and define splice points to be equally spaced along the sample:

```
teoperator drum --slices 16 vocals.wav
```

## Web server ([teoperator.com](https://teoperator.com))

<p align="center">
<a href="https://op1.schollz.com/patch?audioURL=https%3A%2F%2Fcdn.loc.gov%2Fservice%2Fgdc%2Fgdcarpl%2Fgdcarpl-1624415%2F1624415.mp3&secondsStart=982&secondsEnd=1002"><img src="/static/image/example2.png"></a>
</p>


The webserver requires a few more dependencies. You can install them via `apt`:

```
$ sudo apt install imagemagick 
$ sudo add-apt-repository ppa:chris-needham/ppa
$ sudo apt-get update
$ sudo apt-get install audiowaveform
$ sudo -H python3 -m pip install youtube-dl
```

And then you can run the server via

```
$ git clone https://github.com/schollz/teoperator
$ cd teoperator && go build -v
$ teoperator server
```

Then open a browser to `localhost:8053`!


# License

MIT license

Please note THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
