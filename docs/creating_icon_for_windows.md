


### Creating Icon for windows:

To regenerate the .ico file, you can copy the following function into your .bashrc or bash_profile
and then run the command `svgtoico jxwatcher scalable/jxwatcher.svg windows/jxwatcher.ico`
This function uses Inkscape to convert SVG to PNG and ImageMagick to convert PNGs to ICO.

Make sure you have Inkscape and ImageMagick installed:
```
sudo apt install inkscape imagemagick  
```

Copy this to your .bashrc or .bash_profile

```
svgtoico(){
    # $1: Base name for the icon (e.g., "my_icon")
    # $2: Path to the source SVG file (e.g., "path/to/my_icon.svg")
    # $3: Desired path for the output ICO file (e.g., "path/to/output.ico")

    # Create temporary PNG files for different sizes
    inkscape -w 16 -h 16 -o "$1-16.png" "$2"
    inkscape -w 32 -h 32 -o "$1-32.png" "$2"
    inkscape -w 48 -h 48 -o "$1-48.png" "$2"
    inkscape -w 64 -h 64 -o "$1-64.png" "$2"
    inkscape -w 128 -h 128 -o "$1-128.png" "$2"
    inkscape -w 256 -h 256 -o "$1-256.png" "$2"
    inkscape -w 512 -h 512 -o "$1-512.png" "$2"

    # Combine PNGs into an ICO file
    convert "$1-16.png" "$1-32.png" "$1-48.png" "$1-64.png" "$1-128.png" "$1-256.png" "$1-512.png" "$1.ico"

    # Clean up temporary PNG files
    rm "$1-16.png" "$1-32.png" "$1-48.png" "$1-64.png" "$1-128.png" "$1-256.png" "$1-512.png"

    # Move the generated ICO to the desired location
    mv "$1.ico" "$3"
}
```

you might need to refresh:
```
source ~/.bashrc
```