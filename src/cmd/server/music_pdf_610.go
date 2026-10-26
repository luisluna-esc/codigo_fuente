package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"sgl.local/sistema-gestion-liderazgo/internal/qrcode"
	"sort"
	"strconv"
	"strings"
	"time"
)

// A small PDF writer keeps generated requirement documents available on the
// server, including when the applicant has not opened the certificates screen.
type musicPDF610 struct {
	objects             []string
	pages               []int
	images              map[string]int
	fonts613            map[string]int
	fontMetrics613      map[string]musicSFNT613
	opacityResources618 map[string]int
	openType613         bool
}

func newMusicPDF610() *musicPDF610 {
	p := &musicPDF610{objects: []string{"", "", "", "", "", ""}, images: map[string]int{}, fonts613: map[string]int{}, fontMetrics613: map[string]musicSFNT613{}}
	p.fonts610()
	return p
}
func (p *musicPDF610) add(s string) int { p.objects = append(p.objects, s); return len(p.objects) }
func pdfLiteral610(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\', '(', ')':
			b.WriteByte('\\')
			b.WriteRune(r)
		case '\n':
			b.WriteString("\\n")
		case '\r':
		case '—', '–':
			b.WriteByte('-')
		case '’', '‘':
			b.WriteByte('\'')
		case '“', '”':
			b.WriteByte('"')
		case '•':
			b.WriteByte(183)
		default:
			if r >= 32 && r <= 255 {
				b.WriteByte(byte(r))
			} else if r > 255 {
				b.WriteByte('?')
			}
		}
	}
	return "(" + b.String() + ")"
}
func pdfText610(x, y, size float64, text string, bold bool) string {
	font := 0
	if bold {
		font = 1
	}
	return pdfTextFont610(x, y, size, text, font)
}
func pdfWrap610(text string, width, size float64) []string {
	return pdfWrapFont610(text, width, size, 0)
}
func pdfWrapFont610(text string, width, size float64, font int) []string {
	if width <= 0 {
		width = 1
	}
	out := []string{}
	for _, paragraph := range strings.Split(text, "\n") {
		line := ""
		for _, word := range strings.Fields(paragraph) {
			test := word
			if line != "" {
				test = line + " " + word
			}
			if line != "" && pdfMeasure610(test, size, font) > width {
				out = append(out, line)
				line = ""
			}
			for pdfMeasure610(word, size, font) > width {
				part := ""
				remaining := []rune(word)
				for len(remaining) > 0 && pdfMeasure610(part+string(remaining[0]), size, font) <= width {
					part += string(remaining[0])
					remaining = remaining[1:]
				}
				if part == "" {
					part = string(remaining[0])
					remaining = remaining[1:]
				}
				out = append(out, part)
				word = string(remaining)
			}
			if line != "" {
				line += " "
			}
			line += word
		}
		out = append(out, line)
	}
	return out
}
func pdfCenter610(y, size float64, text string) string {
	for pdfMeasure610(text, size, 1) > 730 && size > 8 {
		size--
	}
	return pdfText610(421-pdfMeasure610(text, size, 1)/2, y, size, text, true)
}

func (p *musicPDF610) image(source string) int {
	if id := p.images[source]; id != 0 {
		return id
	}
	parts := strings.SplitN(source, ",", 2)
	if len(parts) != 2 {
		return 0
	}
	raw, e := base64.StdEncoding.DecodeString(parts[1])
	if e != nil {
		return 0
	}
	if id := p.jpegOriginal612(raw); id != 0 {
		p.images[source] = id
		return id
	}
	v, valid := musicImagePixels621(raw)
	if !valid {
		return 0
	}
	w, h := v.w, v.h
	mask := ""
	if v.alpha != "" {
		a := v.alpha
		id := p.add(fmt.Sprintf("<< /Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace /DeviceGray /BitsPerComponent 8 /Filter /FlateDecode /Length %d >>\nstream\n%sendstream", w, h, len(a), a))
		mask = fmt.Sprintf(" /SMask %d 0 R", id)
	}
	b := v.rgb
	id := p.add(fmt.Sprintf("<< /Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /FlateDecode /Length %d%s >>\nstream\n%sendstream", w, h, len(b), mask, b))
	p.images[source] = id
	return id
}
func (p *musicPDF610) page(w, h float64, content string, images map[string]int) {
	stream := p.add(fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(content), content))
	x := ""
	for name, id := range images {
		x += fmt.Sprintf("/%s %d 0 R ", name, id)
	}
	fontResources := ""
	for name, id := range p.fonts613 {
		fontResources += fmt.Sprintf("/%s %d 0 R ", name, id)
	}
	opacityResources := ""
	for name, id := range p.opacityResources618 {
		opacityResources += fmt.Sprintf("/%s %d 0 R ", name, id)
	}
	id := p.add(fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.2f %.2f] /Resources << /Font << /F1 3 0 R /F2 4 0 R /F3 5 0 R /F4 6 0 R %s >> /XObject << %s >> /ExtGState << %s >> >> /Contents %d 0 R >>", w, h, fontResources, x, opacityResources, stream))
	p.pages = append(p.pages, id)
}
func (p *musicPDF610) bytes() []byte {
	p.objects[0] = "<< /Type /Catalog /Pages 2 0 R >>"
	kids := ""
	for _, id := range p.pages {
		kids += fmt.Sprintf("%d 0 R ", id)
	}
	p.objects[1] = fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", kids, len(p.pages))
	var b bytes.Buffer
	if p.openType613 {
		b.WriteString("%PDF-1.6\n")
	} else {
		b.WriteString("%PDF-1.4\n")
	}
	b.WriteString("%\xe2\xe3\xcf\xd3\n")
	offsets := []int{0}
	for i, obj := range p.objects {
		offsets = append(offsets, b.Len())
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, n := range offsets[1:] {
		fmt.Fprintf(&b, "%010d 00000 n \n", n)
	}
	fmt.Fprintf(&b, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", len(offsets), xref)
	return b.Bytes()
}

func certificateDetails610(c map[string]any) []string {
	r := legacyMap(c["registration"])
	ev := legacyMap(c["event"])
	if stringValue(c["kind"]) == "ministry" {
		return []string{"Ministerio: " + stringValue(r["name"]), "Iglesia: " + stringValue(r["church"]) + " · Distrito: " + stringValue(r["district"]), "Región: " + stringValue(r["region"]) + " · Misión: " + stringValue(r["mission"]), "Unión: " + stringValue(r["union"]) + " · Voto de Junta: " + stringValue(r["boardVote"]), "Director: " + stringValue(r["director"]) + " · Creación: " + stringValue(r["createdDate"])}
	}
	if c["eventId"] != "" && ev != nil {
		return []string{"Participó en " + stringValue(ev["name"]), stringValue(ev["date"]) + " · " + stringValue(ev["place"]), "Asistencia confirmada"}
	}
	if c["kind"] == "final" {
		return []string{"Completó el 100% de los requisitos aprobados de la Tarjeta Líder de Música.", stringValue(c["church"]) + " · " + stringValue(c["mission"])}
	}
	return []string{"Ha cumplido el programa de formación indicado.", stringValue(c["church"]) + " · " + stringValue(c["mission"])}
}

func musicCertificateBytes610(c, t map[string]any) []byte {
	p := newMusicPDF610()
	var b strings.Builder
	images := map[string]int{}
	pw, ph := musicPaper617(t)
	ux, uy := pw/100, ph/100
	center := func(y, size float64, text string) string {
		for pdfMeasure610(text, size, 1) > pw-112 && size > 8 {
			size--
		}
		return pdfText610(pw/2-pdfMeasure610(text, size, 1)/2, y*ph/595, size, text, true)
	}
	fmt.Fprintf(&b, "1 1 1 rg 0 0 %g %g re f\n", pw, ph)
	if id := p.image(stringValue(t["dataUrl"])); id != 0 {
		images["BG"] = id
		fmt.Fprintf(&b, "q %s%g 0 0 %g 0 0 cm /BG Do Q\n", p.opacity618(t["backgroundOpacity618"]), pw, ph)
	} else if t["backgroundRemoved613"] != true {
		fmt.Fprintf(&b, "0.07 0.18 0.29 RG 3 w 25 25 %g %g re S\n0.70 0.55 0.27 RG 0.8 w 33 33 %g %g re S\n", pw-50, ph-50, pw-66, ph-66)
	}
	for i, raw := range arrayField(t, "images610") {
		im := legacyMap(raw)
		if id := p.image(stringValue(im["dataUrl"])); id != 0 {
			name := "Logo" + strconv.Itoa(i)
			images[name] = id
			b.WriteString(musicImageCommands619(im, t, name, p.opacity618(im["opacity618"])))
		}
	}
	details := certificateDetails610(c)
	name := stringValue(c["fullName"])
	if name == "" {
		name = stringValue(legacyMap(c["registration"])["name"])
	}
	dept := stringValue(t["signatureDepartment"])
	if dept == "" {
		dept = "Firma Departamental"
	}
	coord := stringValue(t["signatureCoordinator"])
	if coord == "" {
		coord = "Firma Coord. General Música"
	}
	date := stringValue(c["issuedAt"])
	if c["kind"] == "final" && c["approvalDate618"] != nil {
		date = stringValue(c["approvalDate618"])
	}
	vars := map[string]string{"nombre": name, "titulo": stringValue(c["title"]), "datos": strings.Join(details, "\n"), "fecha": musicDate619(date), "lugar": stringValue(legacyMap(c["event"])["place"]), "ministerio": stringValue(legacyMap(c["registration"])["name"]), "firma_departamental": dept, "firma_coordinador": coord, "investidura": stringValue(legacyMap(c["investiture"])["date"])}
	vote := strings.TrimSpace(stringValue(legacyMap(c["registration"])["boardVote"]))
	vars["voto_junta"] = vote
	for k, v := range musicRegistryVars622(c) {
		vars[k] = v
	}
	fields := arrayField(t, "fields")
	hasDate := false
	hasVote := false
	if len(fields) > 0 {
		for _, raw := range fields {
			f := legacyMap(raw)
			b.WriteString("0.07 0.18 0.29 rg\n")
			text := stringValue(f["text"])
			if strings.Contains(strings.ToLower(text), "{fecha}") {
				hasDate = true
			}
			if strings.Contains(text, "{voto_junta}") || strings.Contains(text, "{datos}") {
				hasVote = true
			}
			text = musicReplace618(text, vars)
			if f["runs618"] != nil {
				b.WriteString(p.richText618(f, t, vars))
			} else {
				b.WriteString(p.templateText613(f, text, t))
			}
		}
	}
	if (len(fields) == 0 && t["textOnly620"] != true) || t["legacyBody620"] == true {
		b.WriteString("0.07 0.18 0.29 rg\n")
		b.WriteString(center(523, 14, "SISTEMA DE GESTIÓN DE MÚSICA"))
		b.WriteString(center(462, 32, "CERTIFICADO"))
		b.WriteString(center(426, 17, stringValue(c["title"])))
		size := 28.
		for pdfMeasure610(name, size, 1) > pw-142 && size > 12 {
			size--
		}
		b.WriteString(center(352, size, name))
		y := 310.
		for _, detail := range details {
			for _, line := range pdfWrap610(detail, pw-132, 12) {
				b.WriteString(center(y, 12, line))
				y -= 19
			}
		}
	}
	// Every template has two signature spaces, independent from editable body text.
	if c["kind"] == "ministry" && vote != "" && len(fields) > 0 && !hasVote {
		b.WriteString("0.12 0.16 0.20 rg\n")
		b.WriteString(center(145, 11, "Voto de Junta N.º: "+vote))
	}
	for key, original := range musicSignatures613(t) {
		f := musicSignature619(original)
		label := dept
		if key == "coordinator" {
			label = coord
		}
		if f["runs618"] != nil {
			b.WriteString(p.richText618(f, t, vars))
		} else {
			b.WriteString(p.templateText613(f, label, t))
		}
	}
	if !hasDate {
		b.WriteString(pdfText610(55, 56, 9, "Emitido el "+vars["fecha"], false))
	}
	vf := map[string]any{"x": 35., "y": 96.5, "width": 45., "size": 8., "font": "DejaVu Sans", "align": "left", "color": "#1f2933"}
	for _, k := range []string{"x", "y"} {
		if value := mapField(t, "verification618")[k]; value != nil {
			vf[k] = value
		}
	}
	b.WriteString(p.templateText613(vf, "Código de verificación\n"+stringValue(c["token"]), t))
	if c["revokedAt"] != nil {
		b.WriteString("0.65 0.10 0.10 rg\n")
		b.WriteString(center(24, 10, "HISTORIAL · CERTIFICADO NO VIGENTE"))
	}
	verification := stringValue(c["verificationURL610"])
	if verification == "" {
		verification = "MUSICA:" + stringValue(c["token"])
	}
	if t["showQR617"] != false {
		if qr, e := qrcode.New(verification, qrcode.Medium); e == nil {
			bits := qr.Bitmap()
			size := 64.
			x, y := numericValue(t["qrX"])*ux, ph-numericValue(t["qrY"])*uy-size
			if t["qrX"] == nil {
				x = pw * 700 / 842
			}
			if t["qrY"] == nil {
				y = 60
			}
			if x+size > pw {
				x = pw - size
			}
			if y < 0 {
				y = 0
			}
			unit := size / float64(len(bits))
			b.WriteString(fmt.Sprintf("1 1 1 rg %.3f %.3f %.3f %.3f re f\n0 0 0 rg\n", x, y, size, size))
			for row, line := range bits {
				for col, black := range line {
					if black {
						b.WriteString(fmt.Sprintf("%.3f %.3f %.3f %.3f re f\n", x+float64(col)*unit, y+size-float64(row+1)*unit, unit+.015, unit+.015))
					}
				}
			}
		}
	}
	p.page(pw, ph, b.String(), images)
	return p.bytes()
}

func (s *Server) musicCertificatePDF610(w http.ResponseWriter, r *http.Request, db, actor map[string]any) {
	d := musicData550(db)
	scopeDB, scopeActor := db, actor
	if year := r.URL.Query().Get("year"); year != "" {
		arch := legacyMap(mapField(d, "archiveData610")[year])
		old := legacyMap(arch["database"])
		if old == nil {
			legacyError(w, 403, "Gestión no disponible.", nil)
			return
		}
		scopeDB = old
		scopeActor = musicArchiveActor610(old, actor)
		d = cloneMap610(musicData550(old))
		d["templateVersions610"] = mapField(musicData550(db), "templateVersions610")
	}
	key := r.URL.Query().Get("id")
	tokenValue := r.URL.Query().Get("token")
	var c map[string]any
	for _, raw := range mapField(d, "certificates") {
		candidate := legacyMap(raw)
		if tokenValue != "" && candidate["token"] == tokenValue || tokenValue == "" && candidate["id"] == key {
			c = candidate
			break
		}
	}
	if c == nil && tokenValue != "" {
		for _, raw := range arrayField(d, "certificateHistory610") {
			candidate := legacyMap(raw)
			if candidate["token"] == tokenValue {
				c = candidate
				break
			}
		}
	}
	if mid := r.URL.Query().Get("ministryId"); mid != "" {
		m := legacyMap(mapField(d, "ministries610")[mid])
		if m != nil {
			if issued := legacyMap(mapField(d, "certificates")["ministry-registration:"+mid]); issued != nil {
				c = issued
			} else {
				c = map[string]any{"kind": "ministry", "title": "Registro del Ministerio de Música", "registration": m, "fullName": m["name"], "issuedAt": m["createdAt"], "token": mid, "ownerId": "", "ministryId": mid, "templateSnapshot610": mapField(d, "templates")["ministry"]}
			}
		}
	}
	if c == nil {
		legacyError(w, 404, "Certificado no encontrado.", nil)
		return
	}
	ownerID := stringValue(c["ownerId"])
	if ownerID != "" && ownerID != stringValue(actor["id"]) && !musicLeaders610(legacyRole(actor)) && !musicScope550(scopeDB, scopeActor, legacyUser(scopeDB, ownerID)) {
		legacyError(w, 403, "Certificado fuera de tu alcance.", nil)
		return
	}
	t := legacyMap(mapField(d, "templateVersions610")[stringValue(c["templateVersion610"])])
	if t == nil {
		t = legacyMap(c["templateSnapshot610"])
	}
	if t == nil {
		t = legacyMap(mapField(d, "templates")[stringValue(c["kind"])])
	}
	if ref := stringValue(t["version610"]); ref != "" {
		t = legacyMap(mapField(d, "templateVersions610")[ref])
	}
	// Use the exact persisted document for a current certificate. Revoked copies
	// are rendered with a visible non-current marker while retaining their record.
	var pdf []byte
	if c["revokedAt"] == nil {
		for _, raw := range s.legacyCache.State.Files {
			f := legacyMap(raw)
			if f["id"] == c["fileId610"] && c["fileId610"] != nil {
				parts := strings.SplitN(stringValue(f["dataUrl"]), ",", 2)
				if len(parts) == 2 {
					pdf, _ = base64.StdEncoding.DecodeString(parts[1])
				}
				break
			}
		}
	}
	if len(pdf) == 0 {
		pdf = musicCertificateBytes610(c, musicHydrateFonts613(t, s.legacyCache.State.Files))
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", musicDisposition618(musicFilename618(c, false)))
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(200)
	_, _ = w.Write(pdf)
}

func foldMusic610(s string) string {
	return strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u").Replace(strings.ToLower(strings.TrimSpace(s)))
}
func (s *Server) musicMinistriesPDF610(w http.ResponseWriter, r *http.Request, db map[string]any) {
	rows := []map[string]any{}
	q := foldMusic610(r.URL.Query().Get("q"))
	for _, raw := range mapField(musicData550(db), "ministries610") {
		m := legacyMap(raw)
		if m["deletedAt"] == nil && strings.Contains(foldMusic610(stringValue(m["name"])), q) {
			rows = append(rows, m)
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return foldMusic610(stringValue(rows[i]["name"])) < foldMusic610(stringValue(rows[j]["name"]))
	})
	p := newMusicPDF610()
	var b strings.Builder
	y := 765.
	page := 0
	header := func() {
		page++
		b.Reset()
		b.WriteString("0.07 0.18 0.29 rg\n")
		b.WriteString(pdfText610(45, 795, 20, "Ministerios de Música", true))
		b.WriteString(pdfText610(45, 777, 10, fmt.Sprintf("Registro institucional · %d ministerios · Página %d", len(rows), page), false))
		y = 745
	}
	flush := func() { p.page(595, 842, b.String(), nil) }
	header()
	for _, m := range rows {
		lines := []string{stringValue(m["name"]), stringValue(m["type"]) + " · " + stringValue(m["church"]), "Distrito: " + stringValue(m["district"]) + " · Región: " + stringValue(m["region"]), "Misión: " + stringValue(m["mission"]) + " · Unión: " + stringValue(m["union"]), "Director: " + stringValue(m["director"]) + " · " + stringValue(m["directorPhone"]), "Creación: " + stringValue(m["createdDate"]) + " · Voto de Junta: " + stringValue(m["boardVote"]), fmt.Sprintf("Trayectoria: %d años y %d meses", int(numericValue(m["years"])), int(numericValue(m["months"])))}
		if biography := strings.TrimSpace(stringValue(m["biography"])); biography != "" {
			lines = append(lines, "Reseña: "+biography)
		}
		for i, line := range lines {
			size := 10.
			if i == 0 {
				size = 13
			}
			for _, part := range pdfWrap610(line, 505, size) {
				if y < 55 {
					flush()
					header()
				}
				b.WriteString(pdfText610(45, y, size, part, i == 0))
				y -= size + 5
			}
		}
		y -= 14
	}
	if len(rows) == 0 {
		b.WriteString(pdfText610(45, y, 11, "No hay ministerios que coincidan con la búsqueda.", false))
	}
	flush()
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=Ministerios_de_Musica.pdf")
	_, _ = w.Write(p.bytes())
}

func musicTemplateMutation610(db, actor map[string]any, req musicRequest610) error {
	if !musicLeaders610(legacyRole(actor)) || !musicPermit602(db, actor, "certificates.templates") {
		return fmt.Errorf("Solo el equipo de gestión puede administrar plantillas.")
	}
	d := musicData550(db)
	templates := ensureMapField(d, "templates")
	v := req.Value
	id := strings.TrimSpace(stringValue(v["id"]))
	old := legacyMap(templates[id])
	if req.Action != "template.save" {
		return fmt.Errorf("Acción de plantilla no reconocida.")
	}
	if id == "" {
		id = "template_" + token(10)
	} else if old == nil {
		return fmt.Errorf("La plantilla ya no existe.")
	}
	if old != nil && stringValue(v["updatedAt"]) != stringValue(old["updatedAt"]) {
		return fmt.Errorf("La plantilla cambió. Cierra el editor y vuelve a abrirla antes de guardar.")
	}
	name := strings.TrimSpace(stringValue(v["name"]))
	if name == "" || len([]rune(name)) > 120 {
		return fmt.Errorf("Indica un nombre de plantilla de hasta 120 caracteres.")
	}
	out := map[string]any{"id": id, "name": name, "certificateType": stringValue(v["certificateType"]), "updatedAt": time.Now().UTC().Format(time.RFC3339Nano)}
	if err := musicTemplateSettings618(v, out); err != nil {
		return err
	}
	if err := musicPaperSettings617(v, out); err != nil {
		return err
	}
	out["designLocked622"] = v["designLocked622"] == true
	out["registryLayout622"] = v["registryLayout622"] == true
	out["registryLayout623"] = v["registryLayout623"] == true
	if v["textOnly620"] == true {
		out["textOnly620"] = true
		if v["legacyBody620"] == true {
			out["legacyBody620"] = true
		}
	} else {
		for _, key := range []string{"signatureDepartment", "signatureCoordinator"} {
			value := strings.TrimSpace(stringValue(v[key]))
			if value == "" || len([]rune(value)) > 400 {
				return fmt.Errorf("Completa los dos rótulos de firma (hasta 400 caracteres).")
			}
			out[key] = value
		}
	}
	if source := stringValue(v["dataUrl"]); source != "" {
		if !validMusicPhoto592(source) {
			return fmt.Errorf("Usa un fondo PNG/JPG válido de hasta 2 MB.")
		}
		out["dataUrl"] = source
		out["width"] = v["width"]
		out["height"] = v["height"]
	}
	if v["fields"] != nil {
		fields, ok := musicTemplateFields592(v["fields"])
		if !ok {
			return fmt.Errorf("Revisa los textos y posiciones de la plantilla.")
		}
		out["fields"] = fields
	}
	images := []any{}
	if len(arrayField(v, "images610")) > 10 {
		return fmt.Errorf("Una plantilla admite hasta diez imágenes.")
	}
	for _, raw := range arrayField(v, "images610") {
		im := legacyMap(raw)
		if !validMusicPhoto592(stringValue(im["dataUrl"])) {
			return fmt.Errorf("Usa un logo PNG/JPG válido.")
		}
		x, y, w, h := numericValue(im["x"]), numericValue(im["y"]), numericValue(im["width"]), numericValue(im["height"])
		if x < 0 || y < 0 || w < 1 || h < 1 || x+w > 100.01 || y+h > 100.01 {
			return fmt.Errorf("Coloca las imágenes dentro de la plantilla.")
		}
		opacity, valid := musicNumber618(im, "opacity618", 0, 1, 1)
		if !valid {
			return fmt.Errorf("Revisa la opacidad de la imagen.")
		}
		rotation, valid := musicNumber618(im, "rotation619", -180, 180, 0)
		image := map[string]any{"dataUrl": im["dataUrl"], "x": x, "y": y, "width": w, "height": h, "opacity618": opacity, "rotation619": rotation}
		if !valid || !musicImageBounds619(image, out) {
			return fmt.Errorf("Coloca la imagen girada dentro de la hoja.")
		}
		images = append(images, image)
	}
	out["images610"] = images
	out["backgroundRemoved613"] = v["backgroundRemoved613"] == true
	if v["textOnly620"] != true && v["signatures613"] != nil {
		signatures := map[string]any{}
		for _, key := range []string{"department", "coordinator"} {
			f := cloneMap610(legacyMap(mapField(v, "signatures613")[key]))
			f["text"] = key
			if f["runs618"] != nil {
				if key == "department" {
					f["text"] = out["signatureDepartment"]
				} else {
					f["text"] = out["signatureCoordinator"]
				}
			}
			clean, valid := musicTemplateFields592([]any{f})
			if !valid {
				return fmt.Errorf("Revisa las posiciones y estilos de las firmas.")
			}
			signatures[key] = clean[0]
		}
		out["signatures613"] = signatures
	}
	for _, key := range []string{"qrX", "qrY", "nameX", "nameY", "titleY", "detailsY"} {
		if v[key] != nil {
			n := numericValue(v[key])
			if n < 0 || n > 100 {
				return fmt.Errorf("Posición de plantilla inválida.")
			}
			out[key] = n
		}
	}
	musicTextOnly620(out)
	templates[id] = out
	return nil
}
