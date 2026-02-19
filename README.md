# remove-exif-gpsinfo

This service removes GPS EXIF tags from JPEG images. It runs as an HTTP server and exposes a single endpoint:

```
POST /remove-gps
Content-Type: multipart/form-data
field name: image  (the JPEG file)
```

On success the response body contains the (possibly modified) JPEG bytes with
`Content-Type: image/jpeg`.

* If no GPS tags are present (or there is no EXIF data) the server
  doesn't send the image back – a `204 No Content` status is returned with
  `X-GPS-Removed: false` and additional `X-Has-EXIF`/`X-Has-GPS` headers.
  The client already has the original bytes so re‑sending them would be a
  waste of bandwidth.
* When GPS data is removed the response body contains the modified JPEG
  bytes with `Content-Type: image/jpeg` and `X-GPS-Removed: true`.

## Running

The server listens on the port specified by the `PORT` environment variable.
If `PORT` is unset it defaults to `8080`.

```sh
# start on default port
PORT=8080 go run .

# or override
PORT=9090 go run .
```

## Example curl

```
curl — a command-line tool for making HTTP requests.

-v — "verbose" mode, prints detailed request/response info to the terminal (headers, connection details, etc.).

-F "image=@gps-img.jpg" — sends the file gps-img.jpg as a multipart form upload. The @ symbol tells curl to read from a file. The field name is image, so on the server side you'd access it via $_FILES['image'].

http://localhost:8080/remove-gps — the endpoint being hit. A server is running locally on port 8080, and there's a route /remove-gps that presumably strips GPS/EXIF metadata from the uploaded image.

> out.jpg — redirects the response body (the processed image returned by the server) into a new file called out.jpg.
```

```sh
curl -v -F "image=@gps-img.jpg" http://localhost:8080/remove-gps > out2.jpg
```

`out.jpg` will be the same as `gps-img.jpg` except that any GPS EXIF tags have
been removed. If the image contains no GPS data the file is identical to the
upload.

## Error responses

* `400 Bad Request` – client errors such as missing `image` field or a
  non-JPEG payload.
* `500 Internal Server Error` – unexpected conditions while processing the
  image.

## Example output from curl

```sh
* Host localhost:8080 was resolved.
* IPv6: ::1
* IPv4: 127.0.0.1
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0*   Trying [::1]:8080...
* Connected to localhost (::1) port 8080
> POST /remove-gps HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/8.5.0
> Accept: */*
> Content-Length: 161915
> Content-Type: multipart/form-data; boundary=------------------------iegIYHInUcoFVWLeG5UopP
> 
} [65536 bytes data]
* We are completely uploaded and fine
< HTTP/1.1 200 OK
< Content-Type: image/jpeg
< X-Gps-Removed: true
< Date: Thu, 19 Feb 2026 17:59:51 GMT
< Transfer-Encoding: chunked
< 
{ [3958 bytes data]
100  315k    0  157k  100  158k  16.5M  16.6M --:--:-- --:--:-- --:--:-- 34.2M
* Connection #0 to host localhost left intact
```