package utils

import (
	"bytes"
	"fmt"
	"github.com/puzpuzpuz/xsync/v3"
	"text/template"
)

var templateCache = xsync.NewMapOf[string, *template.Template]()

type sharedContext struct {
	Domain string
}

func getTemplateFromCache(tpl string) (*template.Template, error) {
	var err error
	t, _ := templateCache.Compute(tpl, func(oldValue *template.Template, loaded bool) (newValue *template.Template, delete bool) {
		if loaded && oldValue != nil {
			return oldValue, false
		}
		newValue, err = template.New("").Parse(tpl)
		if err != nil {
			newValue = nil
		}
		return
	})
	return t, err
}

func RenderSharedTemplate(tpl, domain string) (string, error) {
	t, err := getTemplateFromCache(tpl)

	if err != nil {
		return "", fmt.Errorf("template parse: %w", err)
	}

	buf := &bytes.Buffer{}

	err = t.Execute(buf, sharedContext{Domain: domain})
	if err != nil {
		return "", fmt.Errorf("template exec: %w", err)
	}

	return buf.String(), nil
}
