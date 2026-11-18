package storage

import "testing"

func TestSniffImage(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		want string
	}{
		{"jpeg", []byte("\xff\xd8\xff\xe0\x00\x10JFIF"), "image/jpeg"},
		{"png", []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"), "image/png"},
		{"gif", []byte("GIF89a\x01\x00\x01\x00"), "image/gif"},
		{"webp", []byte("RIFF\x24\x00\x00\x00WEBPVP8 "), "image/webp"},
		{"avif", []byte("\x00\x00\x00\x1cftypavif\x00\x00\x00\x00avifmif1"), "image/avif"},
		{"avif sequence", []byte("\x00\x00\x00\x1cftypavis\x00\x00\x00\x00avismsf1"), "image/avif"},
	}
	for _, c := range cases {
		got := sniffImage(c.data)
		if got != c.want {
			t.Errorf("%s: sniffImage = %q, want %q", c.name, got, c.want)
		}
		if _, ok := allowed[got]; !ok {
			t.Errorf("%s: %q should be allowed", c.name, got)
		}
	}
}

func TestSniffImageRejectsOtherFiles(t *testing.T) {
	rejected := map[string][]byte{
		"svg":        []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`),
		"html":       []byte("<!DOCTYPE html><html><body>hi</body></html>"),
		"text":       []byte("just some text"),
		"heic":       []byte("\x00\x00\x00\x18ftypheic\x00\x00\x00\x00mif1heic"),
		"pdf":        []byte("%PDF-1.7"),
		"empty":      {},
		"short mp4":  []byte("\x00\x00\x00\x18ftypmp42"),
		"renamed js": []byte("alert(document.cookie)"),
	}
	for name, data := range rejected {
		got := sniffImage(data)
		if _, ok := allowed[got]; ok {
			t.Errorf("%s: sniffed as %q, which would be accepted", name, got)
		}
	}
}
