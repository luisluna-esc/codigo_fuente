package main

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http/httptest"
	"testing"
)

func Test621CachedArtworkIsLosslessAndBounded(t *testing.T) {
	musicImageCache621.Lock()
	musicImageCache621.values = map[[32]byte]musicPDFImage621{}
	musicImageCache621.order = nil
	musicImageCache621.bytes = 0
	musicImageCache621.Unlock()
	var first musicPDFImage621
	for i := 0; i < 12; i++ {
		im := image.NewNRGBA(image.Rect(0, 0, 2, 2))
		im.SetNRGBA(0, 0, color.NRGBA{uint8(i), 41, 83, 255})
		im.SetNRGBA(1, 0, color.NRGBA{5, 90, 211, 123})
		im.SetNRGBA(0, 1, color.NRGBA{222, 111, 3, 0})
		im.SetNRGBA(1, 1, color.NRGBA{255, 255, 255, 255})
		var raw bytes.Buffer
		_ = png.Encode(&raw, im)
		v, valid := musicImagePixels621(raw.Bytes())
		if !valid || v.w != 2 || v.h != 2 {
			t.Fatal("original dimensions changed")
		}
		decode := func(s string) []byte {
			z, err := zlib.NewReader(bytes.NewBufferString(s))
			if err != nil {
				t.Fatal(err)
			}
			defer z.Close()
			b, _ := io.ReadAll(z)
			return b
		}
		if !bytes.Equal(decode(v.rgb), []byte{uint8(i), 41, 83, 5, 90, 211, 222, 111, 3, 255, 255, 255}) || !bytes.Equal(decode(v.alpha), []byte{255, 123, 0, 255}) {
			t.Fatal("pixels or transparency changed")
		}
		cached, _ := musicImagePixels621(raw.Bytes())
		if cached != v {
			t.Fatal("cache changed the image")
		}
		if i == 0 {
			first = v
		} else if v.rgb == first.rgb {
			t.Fatal("changed artwork used stale pixels")
		}
	}
	musicImageCache621.Lock()
	defer musicImageCache621.Unlock()
	if len(musicImageCache621.values) != 8 || musicImageCache621.bytes > musicImageCacheLimit621 {
		t.Fatal("unbounded cache")
	}
}

type musicUnrelated621 struct{ calls *int }

func (s musicUnrelated621) MarshalJSON() ([]byte, error) {
	*s.calls++
	return []byte(`{"unrelated":true}`), nil
}

func Test621PreviewValidatesOnlyRequestedTemplate(t *testing.T) {
	s, db := fixture550(t)
	musicEnsure610(db)
	st, _ := s.loadLegacy()
	calls := 0
	db["unrelated621"] = musicUnrelated621{&calls}
	before, _ := json.Marshal(mapField(musicData550(db), "templates"))
	v := certificateTemplate617()
	v["id"] = "existing-is-previewed-without-saving"
	r := httptest.NewRequest("POST", "/api/music/v610", nil)
	w := httptest.NewRecorder()
	if !s.musicPreview617(w, r, st, db, legacyUser(db, "admin"), musicRequest610{Action: "template.preview", Value: v, RestoreGeneration: legacyCanonicalRestoreGeneration(st.State)}) || w.Code != 200 {
		t.Fatal("preview failed", w.Code, w.Body.String())
	}
	after, _ := json.Marshal(mapField(musicData550(db), "templates"))
	if calls != 0 || !bytes.Equal(before, after) {
		t.Fatal("preview copied unrelated data or modified templates")
	}
	if !bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF")) {
		t.Fatal("missing PDF")
	}
}

func Test621BundledFontReferencesNeedNoUpload(t *testing.T) {
	fonts := musicScanFonts617(nil, true)
	if len(fonts) == 0 {
		t.Fatal("no bundled fonts")
	}
	for _, f := range fonts {
		if f.AssetID621 == "" || musicFontFile613(nil, f.AssetID621) == nil {
			t.Fatal("bundled reference unavailable", f.Family)
		}
	}
}
