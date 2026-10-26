package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type icoEntry struct {
	Width, Height, Colors, Reserved byte
	Planes, BitCount                uint16
	Size, Offset                    uint32
	Data                            []byte
}

type section struct {
	Name                                            string
	VirtualSize, VirtualAddress, RawSize, RawOffset uint32
}

func align(v, a uint32) uint32 {
	if a == 0 {
		return v
	}
	return (v + a - 1) &^ (a - 1)
}
func u16(b []byte, o int) uint16    { return binary.LittleEndian.Uint16(b[o : o+2]) }
func u32(b []byte, o int) uint32    { return binary.LittleEndian.Uint32(b[o : o+4]) }
func p16(b []byte, o int, v uint16) { binary.LittleEndian.PutUint16(b[o:o+2], v) }
func p32(b []byte, o int, v uint32) { binary.LittleEndian.PutUint32(b[o:o+4], v) }

func readICO(path string) ([]icoEntry, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	if len(b) < 6 || u16(b, 0) != 0 || u16(b, 2) != 1 {
		return nil, errors.New("ICO no válido")
	}
	n := int(u16(b, 4))
	if n < 1 || len(b) < 6+n*16 {
		return nil, errors.New("ICO incompleto")
	}
	out := make([]icoEntry, 0, n)
	for i := 0; i < n; i++ {
		o := 6 + i*16
		sz := u32(b, o+8)
		off := u32(b, o+12)
		if int(off+sz) > len(b) {
			return nil, errors.New("imagen ICO fuera de rango")
		}
		out = append(out, icoEntry{b[o], b[o+1], b[o+2], b[o+3], u16(b, o+4), u16(b, o+6), sz, off, append([]byte(nil), b[off:off+sz]...)})
	}
	return out, nil
}

func buildResource(entries []icoEntry, baseRVA uint32) []byte {
	n := len(entries)
	root := 0
	iconType := 32
	iconLangStart := iconType + 16 + n*8
	groupType := iconLangStart + n*24
	groupLang := groupType + 24
	dataStart := groupLang + 24
	dataEntries := n + 1
	blobStart := align(uint32(dataStart+dataEntries*16), 4)
	groupSize := 6 + n*14
	total := blobStart
	for _, e := range entries {
		total = align(total, 4) + e.Size
	}
	groupOffset := align(total, 4)
	total = groupOffset + uint32(groupSize)
	b := make([]byte, total)
	// root directory: RT_ICON=3, RT_GROUP_ICON=14
	p16(b, root+14, 2)
	p32(b, root+16, 3)
	p32(b, root+20, 0x80000000|uint32(iconType))
	p32(b, root+24, 14)
	p32(b, root+28, 0x80000000|uint32(groupType))
	// icon type directory
	p16(b, iconType+14, uint16(n))
	for i := 0; i < n; i++ {
		eoff := iconType + 16 + i*8
		ld := iconLangStart + i*24
		p32(b, eoff, uint32(i+1))
		p32(b, eoff+4, 0x80000000|uint32(ld))
		p16(b, ld+14, 1)
		p32(b, ld+16, 0x409)
		p32(b, ld+20, uint32(dataStart+i*16))
	}
	// group type directory, resource id 1, en-US
	p16(b, groupType+14, 1)
	p32(b, groupType+16, 1)
	p32(b, groupType+20, 0x80000000|uint32(groupLang))
	p16(b, groupLang+14, 1)
	p32(b, groupLang+16, 0x409)
	p32(b, groupLang+20, uint32(dataStart+n*16))
	// actual icon blobs and data entries
	cursor := blobStart
	for i, e := range entries {
		cursor = align(cursor, 4)
		copy(b[cursor:cursor+e.Size], e.Data)
		de := dataStart + i*16
		p32(b, de, baseRVA+cursor)
		p32(b, de+4, e.Size)
		cursor += e.Size
	}
	// group icon bytes
	cursor = groupOffset
	p16(b, int(cursor), 0)
	p16(b, int(cursor)+2, 1)
	p16(b, int(cursor)+4, uint16(n))
	o := int(cursor) + 6
	for i, e := range entries {
		b[o] = e.Width
		b[o+1] = e.Height
		b[o+2] = e.Colors
		b[o+3] = e.Reserved
		p16(b, o+4, e.Planes)
		p16(b, o+6, e.BitCount)
		p32(b, o+8, e.Size)
		p16(b, o+12, uint16(i+1))
		o += 14
	}
	de := dataStart + n*16
	p32(b, de, baseRVA+groupOffset)
	p32(b, de+4, uint32(groupSize))
	return b
}

func patch(exePath, icoPath string) error {
	ico, e := readICO(icoPath)
	if e != nil {
		return e
	}
	b, e := os.ReadFile(exePath)
	if e != nil {
		return e
	}
	if len(b) < 0x100 || string(b[:2]) != "MZ" {
		return errors.New("EXE no PE")
	}
	pe := int(u32(b, 0x3c))
	if pe+24 > len(b) || string(b[pe:pe+4]) != "PE\x00\x00" {
		return errors.New("cabecera PE inválida")
	}
	numOff := pe + 6
	num := int(u16(b, numOff))
	optSize := int(u16(b, pe+20))
	opt := pe + 24
	if opt+optSize > len(b) {
		return errors.New("opcional fuera de rango")
	}
	magic := u16(b, opt)
	if magic != 0x20b {
		return fmt.Errorf("solo PE32+ soportado: %x", magic)
	}
	secAlign := u32(b, opt+32)
	fileAlign := u32(b, opt+36)
	sizeHeaders := u32(b, opt+60)
	secTable := opt + optSize
	var secs []section
	maxVA := uint32(0)
	firstRaw := uint32(len(b))
	for i := 0; i < num; i++ {
		o := secTable + i*40
		if o+40 > len(b) {
			return errors.New("tabla de secciones truncada")
		}
		nameBytes := b[o : o+8]
		end := 0
		for end < 8 && nameBytes[end] != 0 {
			end++
		}
		name := string(nameBytes[:end])
		vs := u32(b, o+8)
		va := u32(b, o+12)
		rs := u32(b, o+16)
		ro := u32(b, o+20)
		if name == ".rsrc" {
			// Algunos comandos ya incorporan recursos de Windows durante go build.
			// Tratarlo como estado valido hace que COMPILAR_TODO sea idempotente.
			return nil
		}
		secs = append(secs, section{name, vs, va, rs, ro})
		if ro > 0 && ro < firstRaw {
			firstRaw = ro
		}
		endva := va
		if rs > vs {
			endva += rs
		} else {
			endva += vs
		}
		if endva > maxVA {
			maxVA = endva
		}
	}
	if uint32(secTable+(num+1)*40) > sizeHeaders || uint32(secTable+(num+1)*40) > firstRaw {
		return errors.New("no hay espacio para otra cabecera de sección")
	}
	newVA := align(maxVA, secAlign)
	rsrc := buildResource(ico, newVA)
	rawOff := align(uint32(len(b)), fileAlign)
	rawSize := align(uint32(len(rsrc)), fileAlign)
	out := make([]byte, rawOff+rawSize)
	copy(out, b)
	copy(out[rawOff:], rsrc)
	sh := secTable + num*40
	copy(out[sh:sh+8], []byte(".rsrc\x00\x00\x00"))
	p32(out, sh+8, uint32(len(rsrc)))
	p32(out, sh+12, newVA)
	p32(out, sh+16, rawSize)
	p32(out, sh+20, rawOff)
	p32(out, sh+36, 0x40000040)
	p16(out, numOff, uint16(num+1))
	p32(out, opt+56, align(newVA+uint32(len(rsrc)), secAlign))
	dataDir := opt + 112
	p32(out, dataDir+16, newVA)
	p32(out, dataDir+20, uint32(len(rsrc)))
	p32(out, opt+64, 0)
	tmp := exePath + ".icon.tmp"
	if e = os.WriteFile(tmp, out, 0755); e != nil {
		return e
	}
	return os.Rename(tmp, exePath)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "uso: iconpatch icon.ico exe [exe...]\n")
		os.Exit(2)
	}
	ico := os.Args[1]
	for _, exe := range os.Args[2:] {
		if e := patch(exe, ico); e != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", filepath.Base(exe), e)
			os.Exit(1)
		}
		fmt.Printf("icono aplicado: %s\n", filepath.Base(exe))
	}
}
