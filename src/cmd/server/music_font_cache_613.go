package main

import (
	"crypto/sha256"
	"sync"
)

// Bound the cache: repeated attendance certificates reuse the same compressed
// font programs, without retaining every uploaded font for the server lifetime.
var musicFontCache613 = struct {
	sync.Mutex
	values map[[32]byte]musicSFNT613
	order  [][32]byte
}{values: map[[32]byte]musicSFNT613{}}

func musicCachedFont613(raw []byte) (musicSFNT613, error) {
	key := sha256.Sum256(raw)
	musicFontCache613.Lock()
	value, ok := musicFontCache613.values[key]
	musicFontCache613.Unlock()
	if ok {
		return value, nil
	}
	font, err := parseMusicSFNT613(raw)
	if err != nil {
		return font, err
	}
	musicFontCache613.Lock()
	defer musicFontCache613.Unlock()
	if existing, ok := musicFontCache613.values[key]; ok {
		return existing, nil
	}
	if len(musicFontCache613.order) >= 16 {
		delete(musicFontCache613.values, musicFontCache613.order[0])
		musicFontCache613.order = musicFontCache613.order[1:]
	}
	musicFontCache613.order = append(musicFontCache613.order, key)
	musicFontCache613.values[key] = font
	return font, nil
}
