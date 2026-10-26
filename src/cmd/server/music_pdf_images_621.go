package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"image"
	"image/color"
	"sync"
)

// Cache only losslessly encoded pixels, bounded in both bytes and entry count.
// Edits reuse the original full-resolution background and logo programs.
type musicPDFImage621 struct {
	w, h       int
	rgb, alpha string
}

var musicImageCache621 = struct {
	sync.Mutex
	values map[[32]byte]musicPDFImage621
	order  [][32]byte
	bytes  int
}{values: map[[32]byte]musicPDFImage621{}}

const musicImageCacheLimit621 = 24 << 20

func musicImagePixels621(raw []byte) (musicPDFImage621, bool) {
	key := sha256.Sum256(raw)
	musicImageCache621.Lock()
	v, found := musicImageCache621.values[key]
	musicImageCache621.Unlock()
	if found {
		return v, true
	}
	im, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return v, false
	}
	r := im.Bounds()
	v.w, v.h = r.Dx(), r.Dy()
	if v.w < 1 || v.h < 1 || v.w > 6000 || v.h > 6000 {
		return v, false
	}
	rgb := make([]byte, 0, v.w*v.h*3)
	alpha := make([]byte, 0, v.w*v.h)
	transparent := false
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			c := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
			rgb = append(rgb, c.R, c.G, c.B)
			alpha = append(alpha, c.A)
			transparent = transparent || c.A != 255
		}
	}
	compress := func(data []byte) string {
		var out bytes.Buffer
		z := zlib.NewWriter(&out)
		_, _ = z.Write(data)
		_ = z.Close()
		return out.String()
	}
	v.rgb = compress(rgb)
	if transparent {
		v.alpha = compress(alpha)
	}
	size := len(v.rgb) + len(v.alpha)
	if size > musicImageCacheLimit621 {
		return v, true
	}
	musicImageCache621.Lock()
	defer musicImageCache621.Unlock()
	if previous, ok := musicImageCache621.values[key]; ok {
		return previous, true
	}
	for len(musicImageCache621.order) > 0 && (len(musicImageCache621.order) >= 8 || musicImageCache621.bytes+size > musicImageCacheLimit621) {
		old := musicImageCache621.order[0]
		prior := musicImageCache621.values[old]
		musicImageCache621.bytes -= len(prior.rgb) + len(prior.alpha)
		delete(musicImageCache621.values, old)
		musicImageCache621.order = musicImageCache621.order[1:]
	}
	musicImageCache621.values[key] = v
	musicImageCache621.order = append(musicImageCache621.order, key)
	musicImageCache621.bytes += size
	return v, true
}
