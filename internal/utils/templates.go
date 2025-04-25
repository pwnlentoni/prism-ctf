package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/puzpuzpuz/xsync/v3"
	"regexp"
	"text/template"
	"text/template/parse"
)

var tplFuncMap = template.FuncMap{
	"random": func(num int) string {
		rb := make([]byte, num)
		_, _ = rand.Read(rb) // is guaranteed per doc to not return error
		return hex.EncodeToString(rb)
	},
}

var tplFuncRegexMap = map[string]func([]parse.Node) (string, error){
	"random": func(nodes []parse.Node) (string, error) {
		if len(nodes) != 1 {
			return "", fmt.Errorf("random: expected 1 argument, got %d", len(nodes))
		}
		if nodes[0].Type() != parse.NodeNumber {
			return "", fmt.Errorf("random: expected number, got %d", nodes[0].Type())
		}
		num := nodes[0].(*parse.NumberNode)
		if !num.IsInt {
			return "", fmt.Errorf("random: expected int, got %s", num.Text)
		}
		return fmt.Sprintf("[0-9a-f]{%d}", num.Int64*2), nil
	},
}

func init() {
	// just some sanity checks to make sure we don't forget to add a function to the map
	for k := range tplFuncMap {
		_, ok := tplFuncRegexMap[k]
		if !ok {
			panic("missing regex function for " + k)
		}
	}
	for k := range tplFuncRegexMap {
		_, ok := tplFuncMap[k]
		if !ok {
			panic("regex function specified for non existing function " + k)
		}
	}
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

func FlagRegex(tpl string) (string, error) {
	t, err := compileTemplate(tpl)
	if err != nil {
		return "", err
	}

	reg := ""

	var walk func(n parse.Node) error
	walk = func(n parse.Node) error {
		switch n.Type() {
		case parse.NodeText:
			reg += regexp.QuoteMeta(string(n.(*parse.TextNode).Text))
		case parse.NodeList:
			for _, node := range n.(*parse.ListNode).Nodes {
				err := walk(node)
				if err != nil {
					return err
				}
			}
			break
		case parse.NodeAction:
			nn := n.(*parse.ActionNode)
			for _, cmd := range nn.Pipe.Cmds {
				err := walk(cmd)
				if err != nil {
					return err
				}
			}
		case parse.NodeCommand:
			nn := n.(*parse.CommandNode)
			if nn.Args[0].Type() != parse.NodeIdentifier {
				return fmt.Errorf("expected identifier, got %d", nn.Args[0].Type())
			}
			ident := nn.Args[0].(*parse.IdentifierNode)
			f, ok := tplFuncRegexMap[ident.Ident]
			if !ok {
				return fmt.Errorf("unknown function %s", ident.Ident)
			}
			rr, err := f(nn.Args[1:])
			if err != nil {
				return err
			}
			reg += rr
		default:
			break
		}
		return nil
	}

	err = walk(t.Root)

	if err != nil {
		return "", err
	}

	return reg, nil
}
