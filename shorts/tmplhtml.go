package shorts

import (
	"bytes"
	"html/template"

	"github.com/sirupsen/logrus"
)

func RenderHTML(
	tmpl string,
	object interface{},
) string {
	t := template.New("").Funcs(
		template.FuncMap{
			"AsHTML": func(in interface{}) template.HTML {
				switch data := in.(type) {
				case []byte:
					return template.HTML(data)
				case string:
					return template.HTML(data)
				default:
					logrus.Warn("AsHTML on unrecognized type")
					return template.HTML("<!--unknown-->")
				}
			},
		},
	)
	t, err := t.Parse(tmpl)
	if err != nil {
		return ""
	}
	var buffer bytes.Buffer
	t.Execute(&buffer, object)
	return buffer.String()
}
