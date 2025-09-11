### Converting svg to png:

Make sure you have ImageMagick installed:
```
sudo apt install imagemagick  
```

To convert a scalable svg to a PNG file:

```
convert scalable/jxwatcher.svg -resize 256x256 256x256/jxwatcher.png 
convert scalable/jxwatcher.svg -resize 32x32 32x32/jxwatcher.png
```