
# pics-plz

An experimental Go HTTP server that will convert HEIC files to JPEG on demand.

## Building

Depends on Imagemagick C Library. So you need to follow the [Imagemagick instructions](https://github.com/gographics/imagick#ubuntu--debian):

```
sudo apt-get install libmagickwand-dev
```

## Running

```
go run main.go ./samples
```

