experiments:

- can i manipulate the op-1 meta data directly and still upload (try on 1.aif)?
- can convert 1.aif into a new aif with ffmpeg (11.aif) and splice in the meta data before SND?


```
ffmpeg -i 1.aif -af silencedetect=noise=-30dB:d=0.1 -f null -
```