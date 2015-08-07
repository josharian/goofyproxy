package main

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"net/http"

	"code.google.com/p/graphics-go/graphics"
)

func a2the(body []byte) []byte {
	if bytes.Index(body, []byte("html")) == -1 {
		fmt.Println("no html", bytes.Index(body, []byte("html")))
		return body
	}
	body = bytes.Replace(body, []byte(" a "), []byte(" the "), -1)
	return body
}

func flipImage(body []byte) []byte {
	// Is it an image?
	img, typ, err := image.Decode(bytes.NewReader(body))
	if err != nil {
		return body
	}
	dst := image.NewRGBA(img.Bounds())
	graphics.Rotate(dst, img, &graphics.RotateOptions{Angle: math.Pi})
	var buf bytes.Buffer
	switch typ {
	case "png":
		err = png.Encode(&buf, dst)
	case "jpeg":
		err = jpeg.Encode(&buf, dst, nil)
	case "gif":
		err = gif.Encode(&buf, dst, nil)
	}
	if err != nil || buf.Len() == 0 {
		return body
	}
	body = buf.Bytes()
	return body
}

type proxy struct {
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.String())
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	fmt.Println(len(body))
	if len(body) > 1000 {
		fmt.Println(string(body[:1000]))
	} else {
		fmt.Println(string(body))
	}
	body = flipImage(body)
	body = a2the(body)

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Set(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func main() {
	var p proxy
	http.Handle("/", &p)
	if err := http.ListenAndServe(":9999", nil); err != nil {
		log.Fatal(err)
	}
}
