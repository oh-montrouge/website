package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"
	"time"

	"ohmontrouge/webapp/public"
	"ohmontrouge/webapp/templates"

	"github.com/gobuffalo/buffalo/render"
	"github.com/yuin/goldmark"
)

var r *render.Engine

func init() {
	r = render.New(render.Options{
		HTMLLayout:  "layouts/application.plush.html",
		TemplatesFS: templates.FS(),
		AssetsFS:    public.FS(),
		Helpers: render.Helpers{
			"currentYear": func() int { return time.Now().Year() },
			"fmtAmount":   func(amount float64) string { return fmt.Sprintf("%.2f", amount) },
			"toJSON": func(v interface{}) string {
				b, _ := json.Marshal(v)
				return string(b)
			},
			"itoa":           func(i int64) string { return strconv.FormatInt(i, 10) },
			"markdownToHTML": markdownToHTML,
		},
	})
}

// markdownToHTML converts user-supplied markdown to safe HTML.
// XSS safety: goldmark strips raw HTML by default; WithUnsafe() is never enabled.
func markdownToHTML(s string) template.HTML {
	if s == "" {
		return ""
	}
	var buf bytes.Buffer
	goldmark.Convert([]byte(s), &buf) //nolint:errcheck // bytes.Buffer.Write never errors
	return template.HTML(buf.String())
}
