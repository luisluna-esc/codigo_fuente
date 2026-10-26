package xlsxrows

import (
	"bytes"
	"reflect"
	"testing"
)

func TestWriteReadRoundTrip(t *testing.T) {
	rows := [][]string{
		{"nombre", "correo", "telefono"},
		{"María Pérez", "maria@example.org", "70000001"},
		{"Juan O'Connor", "juan@example.org", ""},
	}
	var buf bytes.Buffer
	if err := Write(&buf, rows, WriteOptions{SheetName: "ASPIRANTES_PERFIL", InputColumns: map[int]bool{}, Widths: []float64{30, 28, 24}, FreezeHeader: true, AutoFilter: true}); err != nil {
		t.Fatal(err)
	}
	got, err := Read(bytes.NewReader(buf.Bytes()), 2<<20)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, rows) {
		t.Fatalf("roundtrip mismatch\nwant=%#v\ngot=%#v", rows, got)
	}
}
