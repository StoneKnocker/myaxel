# Myaxel
Axel is my favorite download tool, myaxel implements the main functions of axel, written in go.

Same as axel, myaxel downloads all the data directly to the destination file.  Does not have to concatenate all the downloaded parts.

## Install
```
$ go get -u github.com/progyoung/myaxel
```

## Usage:
```
Usage: myaxel [options] url

optons:
  -T duration
        timeout (default 30m0s)
  -k    do not verify the SSL certificate
  -o string
        local output file name (default "default")
```