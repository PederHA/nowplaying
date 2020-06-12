# Now Playing

Basic "Now Playing" application for Windows that can be used to log the currently playing Spotify track without using the web API. Useful for overlaying "Now Playing" data on a streaming broadcast. 

The program continually polls the Spotify window for updates until it is manually terminated.

I have only read about 50 pages of The Go Programming Language, so this is probably pretty awful.

## Usage

```
$ .\nowplaying.exe [-n][-t]... [PATH]
  -n int
        Polling interval (sec) (default 1)
  -t string
        Window title to look for (default "Spotify")
  [PATH] string
        Path to log file (default "nowplaying.txt")
```