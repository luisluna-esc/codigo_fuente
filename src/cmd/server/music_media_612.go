package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
)

// Embed JPEG artwork as-is. Decoding it and deflating RGB pixels used to make
// printable certificates much larger without adding image detail. PNG artwork
// still uses the lossless RGB/alpha path in musicPDF610.image.
func (p *musicPDF610) jpegOriginal612(raw []byte) int {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || format != "jpeg" || cfg.Width < 1 || cfg.Height < 1 || cfg.Width > 6000 || cfg.Height > 6000 {
		return 0
	}
	if _, err := jpeg.Decode(bytes.NewReader(raw)); err != nil {
		return 0
	}
	space := "/DeviceRGB"
	switch cfg.ColorModel.Convert(color.RGBA{}).(type) {
	case color.Gray:
		space = "/DeviceGray"
	case color.YCbCr, color.RGBA, color.NRGBA:
	default:
		// In particular, Adobe CMYK JPEG needs color conversion. Keep the
		// established lossless decoded path rather than inverting its colors.
		return 0
	}
	return p.add(fmt.Sprintf("<< /Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace %s /BitsPerComponent 8 /Filter /DCTDecode /Length %d >>\nstream\n%sendstream", cfg.Width, cfg.Height, space, len(raw), raw))
}
