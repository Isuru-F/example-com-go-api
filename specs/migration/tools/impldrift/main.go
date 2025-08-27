package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	goparser "github.com/smacker/go-tree-sitter/golang"
)

type MethodInfo struct {
	File      string `json:"file"`
	Receiver  string `json:"receiver"`
	Name      string `json:"name"`
	ParamsRaw string `json:"paramsRaw"`
	ResultsRaw string `json:"resultsRaw"`
}

type RuleExpr struct {
	File  string `json:"file"`
	Line  int    `json:"line"`
	Left  string `json:"left"`
	Op    string `json:"op"`
	Right string `json:"right"`
}

type Snapshot struct {
	Methods  []MethodInfo        `json:"methods"`
	Rules    map[string][]RuleExpr `json:"rules"`
	Exports  []string            `json:"exports"`
	Consts   map[string]float64  `json:"consts"`
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	sub := os.Args[1]
	switch sub {
	case "extract":
		extractCmd(os.Args[2:])
	case "validate":
		validateCmd(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println("impldrift extract --out tools/baseline/ast-baseline.json")
	fmt.Println("impldrift validate --baseline tools/baseline/ast-baseline.json")
}

func extractCmd(args []string) {
	fs := flag.NewFlagSet("extract", flag.ExitOnError)
	out := fs.String("out", "tools/baseline/ast-baseline.json", "output file")
	dir := fs.String("dir", ".", "source directory to analyze")
	_ = fs.Parse(args)
	shot, err := buildSnapshot(*dir)
	if err != nil { fail(err) }
	if err := writeJSON(*out, shot); err != nil { fail(err) }
	fmt.Printf("wrote baseline to %s\n", *out)
}

func validateCmd(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	baseline := fs.String("baseline", "tools/baseline/ast-baseline.json", "baseline json")
	dir := fs.String("dir", ".", "source directory to analyze")
	_ = fs.Parse(args)
	base, err := readSnapshot(*baseline)
	if err != nil { fail(fmt.Errorf("read baseline: %w", err)) }
	curr, err := buildSnapshot(*dir)
	if err != nil { fail(err) }

	violations := []string{}
	// signature shape validation
	for _, m := range curr.Methods {
		if !strings.HasPrefix(m.ParamsRaw, "(") || !strings.Contains(m.ParamsRaw, "endpoint.HTTPRequest[") {
			violations = append(violations, fmt.Sprintf("%s.%s: second param not endpoint.HTTPRequest[...] (%s)", m.Receiver, m.Name, m.ParamsRaw))
		}
		if !strings.HasPrefix(m.ResultsRaw, "(") || !strings.Contains(m.ResultsRaw, "endpoint.HTTPResponse[") || !strings.HasSuffix(m.ResultsRaw, ", error)") {
			violations = append(violations, fmt.Sprintf("%s.%s: results not (*endpoint.HTTPResponse[...], error) (%s)", m.Receiver, m.Name, m.ResultsRaw))
		}
	}
	// scope: exported method set diff
	currSet := setOf(curr.Exports)
	baseSet := setOf(base.Exports)
	for k := range diff(currSet, baseSet) { // new exports in curr
		violations = append(violations, fmt.Sprintf("new exported method: %s", k))
	}
	// naive rules drift: same count and operator tokens when const-involved or numeric literal
	for svcMeth, exprs := range base.Rules {
		currExprs := curr.Rules[svcMeth]
		if len(exprs) != len(currExprs) {
			violations = append(violations, fmt.Sprintf("rule count changed for %s: %d -> %d", svcMeth, len(exprs), len(currExprs)))
			continue
		}
		for i := range exprs {
			b := exprs[i]
			c := currExprs[i]
			leftB := normalizeSide(b.Left)
			leftC := normalizeSide(c.Left)
			rightB := normNum(base.Consts, normalizeSide(b.Right))
			rightC := normNum(curr.Consts, normalizeSide(c.Right))
			if b.Op != c.Op || rightB != rightC || leftB != leftC {
				violations = append(violations, fmt.Sprintf("rule drift %s[%d]: %s %s %s -> %s %s %s", svcMeth, i, b.Left, b.Op, b.Right, c.Left, c.Op, c.Right))
			}
		}
	}
	if len(violations) > 0 {
		fmt.Println("VALIDATION FAILURES:")
		for _, v := range violations { fmt.Println("- ", v) }
		os.Exit(1)
	}
	fmt.Println("validation OK")
}

func buildSnapshot(dir string) (*Snapshot, error) {
	shot := &Snapshot{Rules: map[string][]RuleExpr{}, Consts: map[string]float64{}}
	// collect consts from rules.go
	if err := parseRulesConsts(filepath.Join(dir, "internal/services/rules.go"), shot); err != nil {
		return nil, err
	}
	// scan service files
	var files []string
	err := filepath.WalkDir(filepath.Join(dir, "internal/services"), func(path string, d fs.DirEntry, err error) error {
		if err != nil { return err }
		if d.IsDir() { return nil }
		if strings.HasSuffix(path, "_test.go") || !strings.HasSuffix(path, ".go") { return nil }
		if strings.HasSuffix(path, "rules.go") { return nil }
		files = append(files, path)
		return nil
	})
	if err != nil { return nil, err }
	sort.Strings(files)

	parser := sitter.NewParser()
	parser.SetLanguage(goparser.GetLanguage())
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil { return nil, err }
		tree := parser.Parse(nil, b)
		root := tree.RootNode()
		extractMethodsAndRules(f, string(b), root, shot)
	}
	// exported names
	for _, m := range shot.Methods {
		if isExported(m.Name) {
			shot.Exports = append(shot.Exports, fmt.Sprintf("%s.%s", trimPtr(m.Receiver), m.Name))
		}
	}
	sort.Strings(shot.Exports)
	return shot, nil
}

func parseRulesConsts(path string, shot *Snapshot) error {
	b, err := os.ReadFile(path)
	if err != nil { return err }
	parser := sitter.NewParser()
	parser.SetLanguage(goparser.GetLanguage())
	tree := parser.Parse(nil, b)
	root := tree.RootNode()
	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		if n.Type() == "const_declaration" {
			// crude parsing: scan text lines in this node
			text := string(b[n.StartByte():n.EndByte()])
			lines := strings.Split(text, "\n")
			for _, ln := range lines {
				ln = strings.TrimSpace(ln)
				if strings.Contains(ln, "=") {
					parts := strings.SplitN(ln, "=", 2)
					name := strings.Fields(strings.TrimSpace(parts[0]))
					if len(name) > 0 {
						id := name[0]
						val := strings.TrimSpace(parts[1])
						val = strings.TrimSuffix(val, ",")
						val = strings.TrimSpace(val)
						if f, perr := parseFloat(val); perr == nil {
							shot.Consts[id] = f
						}
					}
				}
			}
		}
		for i := 0; i < int(n.NamedChildCount()); i++ { walk(n.NamedChild(i)) }
	}
	walk(root)
	return nil
}

func extractMethodsAndRules(file, src string, root *sitter.Node, shot *Snapshot) {
	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		if n.Type() == "method_declaration" {
			// receiver, name, param list, result list
			receiver := textOf(src, n.ChildByFieldName("receiver"))
			name := textOf(src, n.ChildByFieldName("name"))
			params := textOf(src, n.ChildByFieldName("parameters"))
			results := textOf(src, n.ChildByFieldName("result"))
			shot.Methods = append(shot.Methods, MethodInfo{
				File: file,
				Receiver: receiver,
				Name: name,
				ParamsRaw: params,
				ResultsRaw: results,
			})
			// collect binary expressions inside body
			body := n.ChildByFieldName("body")
			if body != nil {
				mkey := fmt.Sprintf("%s.%s", strings.TrimSpace(receiver), name)
				list := shot.Rules[mkey]
				collectBinaryRules(file, src, body, shot.Consts, &list)
				shot.Rules[mkey] = list
			}
		}
		for i := 0; i < int(n.NamedChildCount()); i++ { walk(n.NamedChild(i)) }
	}
	walk(root)
}

func collectBinaryRules(file, src string, node *sitter.Node, consts map[string]float64, out *[]RuleExpr) {
	var walk func(n *sitter.Node)
	isRel := func(op string) bool { return op == "<" || op == "<=" || op == ">" || op == ">=" || op == "==" || op == "!=" }
	isNumericish := func(s string) bool {
		if _, ok := consts[strings.TrimSpace(s)]; ok { return true }
		if _, err := parseFloat(s); err == nil { return true }
		return false
	}
	walk = func(n *sitter.Node) {
		if n.Type() == "binary_expression" {
			left := strings.TrimSpace(textOf(src, n.ChildByFieldName("left")))
			op := strings.TrimSpace(textOf(src, n.ChildByFieldName("operator")))
			right := strings.TrimSpace(textOf(src, n.ChildByFieldName("right")))
			if !isRel(op) { goto NEXT }
			// skip common noise
			if strings.Contains(left, "err") || strings.Contains(right, "err") { goto NEXT }
			if strings.Contains(left, "nil") || strings.Contains(right, "nil") { goto NEXT }
			// keep only numeric threshold related comparisons
			if !(isNumericish(left) || isNumericish(right)) { goto NEXT }
			line := int(n.StartPoint().Row) + 1
			*out = append(*out, RuleExpr{File: file, Line: line, Left: left, Op: op, Right: right})
		}
	NEXT:
		for i := 0; i < int(n.NamedChildCount()); i++ { walk(n.NamedChild(i)) }
	}
	walk(node)
}

func textOf(src string, n *sitter.Node) string {
	if n == nil { return "" }
	return src[n.StartByte():n.EndByte()]
}

func setOf(arr []string) map[string]struct{} {
	m := map[string]struct{}{}
	for _, v := range arr { m[v] = struct{}{} }
	return m
}

func diff(a, b map[string]struct{}) map[string]struct{} {
	m := map[string]struct{}{}
	for k := range a { if _, ok := b[k]; !ok { m[k] = struct{}{} } }
	return m
}

func isExported(name string) bool {
	if name == "" { return false }
	r := rune(name[0])
	return r >= 'A' && r <= 'Z'
}

func trimPtr(recv string) string {
	recv = strings.TrimSpace(recv)
	// e.g. (s *ProductService)
	if strings.HasPrefix(recv, "(") && strings.Contains(recv, "*") && strings.Contains(recv, ")") {
		inside := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(recv, "("), ")"))
		parts := strings.Fields(inside)
		if len(parts) == 2 {
			// parts[1] like *ProductService
			return strings.TrimPrefix(parts[1], "*")
		}
	}
	return recv
}

func parseFloat(s string) (float64, error) {
	// strip underscores, quotes
	s = strings.Trim(s, "`\"')")
	s = strings.TrimSpace(strings.ReplaceAll(s, "_", ""))
	var f float64
	_, err := fmt.Sscan(s, &f)
	return f, err
}

func normalizeSide(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ".Body.", ".")
	return s
}

func normNum(consts map[string]float64, s string) string {
	s = strings.TrimSpace(s)
	if v, ok := consts[s]; ok {
		return fmt.Sprintf("%g", v)
	}
	// literal? try float
	if f, err := parseFloat(s); err == nil {
		return fmt.Sprintf("%g", f)
	}
	return s
}

func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil { return err }
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil { return err }
	return os.WriteFile(path, b, 0o644)
}

func readSnapshot(path string) (*Snapshot, error) {
	b, err := os.ReadFile(path)
	if err != nil { return nil, err }
	var s Snapshot
	if err := json.Unmarshal(b, &s); err != nil { return nil, err }
	return &s, nil
}

func fail(err error) { fmt.Fprintln(os.Stderr, err.Error()); os.Exit(2) }
