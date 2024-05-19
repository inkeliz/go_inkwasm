package parser

import (
	"errors"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/inkeliz/go_inkwasm/bind"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type Parser struct {
	PackagesConfig *packages.Config
	parsed         map[string][]*bind.Function
	visited        map[string]bool
}

func NewParser() *Parser {
	return &Parser{
		PackagesConfig: &packages.Config{
			Mode:  packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps | packages.NeedSyntax | packages.NeedTypes,
			Env:   append(os.Environ(), "GOOS=js", "GOARCH=wasm"),
			Tests: true,
		},
		parsed:  make(map[string][]*bind.Function, 128),
		visited: make(map[string]bool, 128),
	}
}

func (p *Parser) ParsePackages(dir string) (map[bind.Package][]*bind.Function, error) {
	pkgs, err := packages.Load(p.PackagesConfig, dir)
	if err != nil {
		return nil, err
	}

	if len(pkgs) == 0 {
		return nil, errors.New("no package found")
	}

	m := make(map[bind.Package][]*bind.Function, 128)
	for _, pkg := range pkgs {
		if strings.Contains(pkg.PkgPath, ".test") {
			info, err := p.ParsePackage(pkg)
			if err != nil {
				return nil, err
			}
			m[bind.Package{Name: pkg.Name, Path: pkg.PkgPath, Dir: filepath.Dir(pkg.GoFiles[0])}] = info
		} else {
			if err := p.parsePackages(m, pkg); err != nil {
				return nil, err
			}
		}
	}

	return m, nil
}

func (p *Parser) parsePackages(m map[bind.Package][]*bind.Function, pkg *packages.Package) error {
	info, err := p.ParsePackage(pkg)
	if err != nil {
		return err
	}
	if len(info) > 0 {
		m[bind.Package{Name: pkg.Name, Path: pkg.PkgPath, Dir: filepath.Dir(pkg.GoFiles[0])}] = info
	}

	for _, imp := range pkg.Imports {
		if !p.visited[imp.ID] {
			p.visited[imp.ID] = true
			if err := p.parsePackages(m, imp); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Parser) ParsePackage(pkg *packages.Package) ([]*bind.Function, error) {
	if len(pkg.GoFiles) == 0 {
		return nil, nil
	}

	var infos []*bind.Function
	for _, f := range pkg.Syntax {
		b, err := p.ParseFile(pkg.PkgPath, pkg.Fset, f)
		if err != nil {
			return nil, err
		}
		infos = append(infos, b...)
	}

	return infos, nil
}

func (p *Parser) ParseFile(pkg string, fset *token.FileSet, file *ast.File) (b []*bind.Function, err error) {
	var info *bind.Function
	astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		switch x := n.(type) {
		case *ast.Comment:
			if info != nil {
				return true
			}
			info, err = p.parseComment(x.Text)
			if err != nil {
				return false
			}
			if info != nil {
				b = append(b, info)
			}
		case *ast.FuncDecl:
			if info == nil {
				return true
			}
			if info.FunctionGolang, err = p.parseFunction(pkg, x); err != nil {
				return false
			}
			if fset != nil {
				info.File = fset.File(x.Pos()).Name()
				info.IsTest = strings.Contains(info.File, "_test")
				info.Line = fset.PositionFor(x.Pos(), true).Line
			}
			info = nil
		case *ast.TypeSpec:
			if info == nil {
				return true
			}
			info.FunctionGolang.Name = x.Name.Name
			info = nil
		case *ast.StructType:
			if info == nil {
				return true
			}
			if info.FunctionGolang, err = p.parseStruct(pkg, x); err != nil {
				return false
			}
			if fset != nil {
				info.File = fset.File(x.Pos()).Name()
				info.IsTest = strings.Contains(info.File, "_test")
				info.Line = fset.PositionFor(x.Pos(), true).Line
			}
		default:
			if info != nil {
				//	fmt.Printf("%T %v \n", x, x)
			}
		}

		return true
	})

	return b, err
}

type filter func(b byte) (filter, error)

func (p *Parser) parseComment(s string) (b *bind.Function, err error) {
	const prefixLen = len("//inkwasm:")

	if len(s) <= prefixLen || !strings.HasPrefix(s, "//inkwasm:") {
		return b, nil
	}

	var jsName, jsHint strings.Builder
	var hint, function filter

	hint = func(b byte) (filter, error) {
		switch b {
		case ' ', '\t':
			return function, nil
		default:
			jsHint.WriteByte(b)
			return hint, nil
		}
	}
	function = func(b byte) (filter, error) {
		jsName.WriteByte(b)
		return function, nil
	}

	i := prefixLen
	l := len(s)
	for f := hint; f != nil; {
		if i >= l {
			break
		}
		if f, err = f(s[i]); err != nil {
			return nil, err
		}
		i++
	}

	return &bind.Function{
		FunctionJavascript: bind.FunctionJavascript{
			Name: strings.TrimSpace(jsName.String()),
			Hint: bind.Hint(jsHint.String()),
		},
	}, nil
}

func (p *Parser) parseFunction(pkg string, f *ast.FuncDecl) (bind.FunctionGolang, error) {
	b := bind.FunctionGolang{
		Name:      f.Name.Name,
		Arguments: nil,
		Result:    nil,
	}

	if err := p.parseFields(pkg, &b.Arguments, f.Type.Params); err != nil {
		return b, err
	}

	if err := p.parseFields(pkg, &b.Result, f.Type.Results); err != nil {
		return b, err
	}
	return b, nil
}

func (p *Parser) parseStruct(pkg string, f *ast.StructType) (bind.FunctionGolang, error) {
	b := bind.FunctionGolang{
		Arguments: nil,
	}

	if err := p.parseFields(pkg, &b.Arguments, f.Fields); err != nil {
		return b, err
	}

	if len(b.Arguments) > 0 && (b.Arguments[0].Name != "_" || b.Arguments[0].Type != "uint64") {
		return b, errors.New(`first argument of exported struct must be uint64 and named as "_"`)
	}

	return b, nil
}

func (p *Parser) parseFields(pkg string, out *[]bind.Argument, fields *ast.FieldList) error {
	if fields == nil || len(fields.List) == 0 {
		return nil
	}

	for _, p := range fields.List {
		if len(p.Names) == 0 {
			p.Names = []*ast.Ident{{Name: "_"}}
		}
		for _, n := range p.Names {
			*out = append(*out, bind.Argument{Name: n.Name})
			arg := &((*out)[len(*out)-1])

			if p.Tag != nil {
				i := strings.Index(p.Tag.Value, "js:")
				if i > -1 || len(p.Tag.Value) > len("js:")+1 {
					s := p.Tag.Value[i+len("js:")+1:]
					if i = strings.Index(s, `"`); i > -1 {
						arg.Tag = s[:i]
					}
				}
			}

			switch t := p.Type.(type) {
			case *ast.ArrayType:
				if err := parseArray(arg, t); err != nil {
					return err
				}
			case *ast.Ident:
				if t.Name == "Object" && pkg == "github.com/inkeliz/go_inkwasm/inkwasm" {
					t.Name = "inkwasm." + t.Name
				}
				if err := parseIdent(arg, t); err != nil {
					return err
				}
			case *ast.SelectorExpr:
				if err := parseSelector(arg, t); err != nil {
					return err
				}
			case *ast.StarExpr:
				if err := parsePointer(arg, t); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func parseIdent(arg *bind.Argument, t *ast.Ident) error {
	arg.ArgType = bind.ModeStatic
	arg.Type = t.Name
	return nil
}

func parsePointer(arg *bind.Argument, t *ast.StarExpr) error {
	arg.ArgType = bind.ModePointer
	arg.SubType = new(bind.Argument)
	switch tt := t.X.(type) {
	case *ast.Ident:
		return parseIdent(arg.SubType, tt)
	case *ast.SelectorExpr:
		return parseSelector(arg.SubType, tt)
	default:
		return errors.New("invalid format")
	}
}

func parseSelector(arg *bind.Argument, t *ast.SelectorExpr) error {
	arg.ArgType = bind.ModeStatic
	switch tt := t.X.(type) {
	case *ast.Ident:
		defer func() {
			arg.Type = arg.Type + "." + t.Sel.Name
		}()
		return parseIdent(arg, tt)
	default:
		return errors.New("invalid format")
	}
}

func parseArray(arg *bind.Argument, t *ast.ArrayType) error {
	if t.Len == nil {
		arg.ArgType = bind.ModeSlice
	} else {
		arg.ArgType = bind.ModeArray
		l, ok := t.Len.(*ast.BasicLit)
		if !ok {
			return errors.New("invalid array length, the size must be explicit not based on constants")
		}
		arg.Len, _ = strconv.ParseUint(l.Value, 10, 64)
	}
	arg.SubType = new(bind.Argument)
	switch tt := t.Elt.(type) {
	case *ast.Ident:
		return parseIdent(arg.SubType, tt)
	case *ast.SelectorExpr:
		return parseSelector(arg.SubType, tt)
	default:
		return errors.New("array/slice of pointers isn't supported")
	}
}
