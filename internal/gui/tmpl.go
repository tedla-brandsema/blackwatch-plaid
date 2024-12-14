package gui

import (
	"bytes"
	"html/template"
)

func Render(tmpl *template.Template, data any) (template.HTML, error) {
	var b []byte
	buffer := bytes.NewBuffer(b)
	err := tmpl.Execute(buffer, data)
	if err != nil {
		return "", err
	}
	return template.HTML(buffer.String()), nil
}

func Merge(templates ...*template.Template) *template.Template {
	var tpl *template.Template
	for _, t := range templates {
		clone := template.Must(t.Clone())
		if tpl == nil {
			tpl = clone
			continue
		}
		_, _ = tpl.AddParseTree(clone.Name(), clone.Tree)
	}
	return tpl
}

//func NewTemplate(name string, templates ...*template.Template) *template.Template {
//	var tpl *template.Template
//
//	return template.Must(template.New(name).AddParseTree(tmpl.Name(), tmpl.Tree))
//}
