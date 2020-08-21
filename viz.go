package viz

import (
	"bytes"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"
)

type Edge struct{ From, To string }

func (e *Edge) Hash() string {
	return e.From + " -> " + e.To
}

type Viz struct {
	ProjectName string
	ProjectRoot string
	Dot         bool
	MaxDepth    int
	Edges       map[string]Edge
}

func (v *Viz) lazyInit() {
	if v.Edges == nil {
		v.Edges = make(map[string]Edge)
	}

	if v.ProjectName == "" {
		v.ProjectName = ProjectNameFrmGoMod(v.ProjectRoot)
	}
}

func (v *Viz) Parse() error {
	return v.parse(v.ProjectRoot, v.ProjectName, 0)
}

func (v *Viz) parse(root, rootpkg string, depth int) error {
	if v.MaxDepth >= 0 && depth > v.MaxDepth {
		return nil
	}

	v.lazyInit()
	pkgs, err := parser.ParseDir(token.NewFileSet(), root,
		func(info os.FileInfo) bool {
			return !strings.HasSuffix(info.Name(), "_test.go")
		}, parser.ParseComments)
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, imp := range file.Imports {
				imppath := unquote(imp.Path.Value)
				if strings.HasPrefix(imppath, v.ProjectName) {
					edge := Edge{From: trim(rootpkg, v.ProjectName), To: trim(imppath, v.ProjectName)}
					v.Edges[edge.Hash()] = edge
				}
			}
		}
	}

	// 递归解析子目录
	dir, err := ioutil.ReadDir(root)
	if err != nil {
		return err
	}

	for _, subdir := range dir {
		if strings.HasPrefix(subdir.Name(), ".") || !subdir.IsDir() || subdir.Name() == "vendor" {
			continue
		}

		if err := v.parse(path.Join(root, subdir.Name()), path.Join(rootpkg, subdir.Name()), depth+1); err != nil {
			return err
		}
	}

	return nil
}

func unquote(s string) string {
	if s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return ""
}

var defaultTemplate = `{{range $key, $edge := .v.Edges}}{{$edge.From}} {{$edge.To}}
{{end}}`

var dotTemplate = `digraph "{{.v.ProjectName}}" {
	label="{{.v.ProjectName}}";
	rankdir=RL;
	node [shape=Mrecord, style=solid];
{{range $key, $edge := .v.Edges}}	"{{$edge.From}}" -> "{{$edge.To}}";
{{end}}}`

func (v *Viz) Render() string {
	tmpl := defaultTemplate
	if v.Dot {
		tmpl = dotTemplate
	}

	t := template.Must(template.New("viz").Parse(tmpl))
	var buf bytes.Buffer
	if err := t.Execute(&buf, map[string]interface{}{"v": v}); err != nil {
		return err.Error()
	}
	return buf.String()
}

func ProjectNameFrmGoMod(rpath string) string {
	data, err := ioutil.ReadFile(path.Join(rpath, "go.mod"))
	if err != nil {
		return ""
	}

	r := regexp.MustCompile("module (.*?)\n")
	res := r.FindStringSubmatch(string(data))
	if len(res) > 1 {
		return res[1]
	}

	return ""
}

func trim(s, p string) string {
	s = strings.TrimPrefix(s, p)
	s = strings.TrimPrefix(s, "/")
	return s
}
