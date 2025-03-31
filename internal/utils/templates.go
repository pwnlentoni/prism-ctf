package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"github.com/puzpuzpuz/xsync/v3"
	"text/template"
)

var tplFuncMap = template.FuncMap{
	"random": func(num int) string {
		palle := make([]byte, num)
		_, _ = rand.Read(palle) // is guaranteed per doc to not return error
		return hex.EncodeToString(palle)
	},
}

var tplCache = xsync.NewMapOf[string, *template.Template]()

func compileTemplate(tpl string) (*template.Template, error) {
	var err error
	t, _ := tplCache.LoadOrTryCompute(tpl, func() (*template.Template, bool) {
		var t *template.Template
		t, err = template.New("flagTemplate").Funcs(tplFuncMap).Parse(tpl)
		if err != nil {
			return nil, true
		}
		return t, false
	})
	return t, err // t == nil ^ err == nil holds because of LoadOrTryCompute
}

func TemplateFlag(tpl string) (string, error) {
	t, err := compileTemplate(tpl)
	if err != nil {
		return "", err
	}
	var flag bytes.Buffer
	err = t.Execute(&flag, nil)
	if err != nil {
		return "", err
	}
	return flag.String(), nil
}
