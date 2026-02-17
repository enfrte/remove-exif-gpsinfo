# remove-exif-gpsinfo

Vibe coded my first golang app.

Take a jpeg image as an argument and look for any GPS coordinates in the image exif metadata. 

* If the image does not contain exif data, skip the processing. 
* Else if exif data does exist, check for gps coordinates. If there are no gps coordinates, skip the processing.
* Else there is exif data and gps coordinates present, so use github.com/dsoprea/go-exif to remove all the gps tags in the safest possible way with least possible chance of corruption. 
