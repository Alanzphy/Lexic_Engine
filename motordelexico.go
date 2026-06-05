package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var languageByExtension = map[string]string{
	".erl": "erlang",
	".hrl": "erlang",
	".go":  "go",
	".c":   "cpp",
	".cc":  "cpp",
	".cpp": "cpp",
	".cxx": "cpp",
	".h":   "cpp",
	".hh":  "cpp",
	".hpp": "cpp",
	".hxx": "cpp",
}

type tokenDef struct {
	name    string
	pattern string
	regex   *regexp.Regexp
	skip    bool
}

type lexerSpec struct {
	languages map[string][]tokenDef
}

type token struct {
	typ    string
	lexeme string
	line   int
	col    int
}

type fileResult struct {
	index    int
	input    string
	language string
	tokens   []token
	duration time.Duration
	err      error
}

type inputList []string

func (i *inputList) String() string {
	return strings.Join(*i, ",")
}

func (i *inputList) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type sexpParser struct {
	text string
	i    int
}

func newSexpParser(text string) *sexpParser {
	return &sexpParser{text: text}
}

func (p *sexpParser) parse() ([]any, error) {
	var items []any
	for {
		p.skipWSAndComments()
		if p.i >= len(p.text) {
			return items, nil
		}

		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		items = append(items, expr)
	}
}

func (p *sexpParser) skipWSAndComments() {
	for p.i < len(p.text) {
		switch p.text[p.i] {
		case ' ', '\t', '\r', '\n':
			p.i++
		case ';':
			for p.i < len(p.text) && p.text[p.i] != '\n' {
				p.i++
			}
		default:
			return
		}
	}
}

func (p *sexpParser) parseExpr() (any, error) {
	p.skipWSAndComments()
	if p.i >= len(p.text) {
		return nil, errors.New("fin inesperado al parsear expresion-s")
	}

	switch p.text[p.i] {
	case '(':
		p.i++
		var out []any
		for {
			p.skipWSAndComments()
			if p.i >= len(p.text) {
				return nil, errors.New("parentesis sin cerrar en expresion-s")
			}
			if p.text[p.i] == ')' {
				p.i++
				return out, nil
			}

			expr, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			out = append(out, expr)
		}
	case '"':
		return p.parseString()
	case ')':
		return nil, errors.New("parentesis de cierre inesperado")
	default:
		return p.parseSymbol()
	}
}

func (p *sexpParser) parseString() (string, error) {
	if p.text[p.i] != '"' {
		return "", errors.New("se esperaba string")
	}
	p.i++

	var out strings.Builder
	for p.i < len(p.text) {
		ch := p.text[p.i]
		p.i++

		if ch == '"' {
			return out.String(), nil
		}

		if ch != '\\' {
			out.WriteByte(ch)
			continue
		}

		if p.i >= len(p.text) {
			return "", errors.New("escape incompleto en string")
		}
		esc := p.text[p.i]
		p.i++

		switch esc {
		case 'n':
			out.WriteByte('\n')
		case 'r':
			out.WriteByte('\r')
		case 't':
			out.WriteByte('\t')
		case '\\':
			out.WriteByte('\\')
		case '"':
			out.WriteByte('"')
		case '\'':
			out.WriteByte('\'')
		case '0':
			out.WriteByte(0)
		default:
			out.WriteByte(esc)
		}
	}

	return "", errors.New("string sin cerrar")
}

func (p *sexpParser) parseSymbol() (string, error) {
	start := p.i
	for p.i < len(p.text) {
		switch p.text[p.i] {
		case '(', ')', ' ', '\t', '\r', '\n':
			if start == p.i {
				return "", errors.New("simbolo vacio")
			}
			return p.text[start:p.i], nil
		default:
			p.i++
		}
	}
	if start == p.i {
		return "", errors.New("simbolo vacio")
	}
	return p.text[start:p.i], nil
}

func parseSpec(specText string) (*lexerSpec, error) {
	forms, err := newSexpParser(specText).parse()
	if err != nil {
		return nil, err
	}
	if len(forms) != 1 {
		return nil, errors.New("el archivo de especificacion debe tener una unica forma raiz")
	}

	root, ok := forms[0].([]any)
	if !ok || len(root) == 0 || root[0] != "spec" {
		return nil, errors.New("la raiz debe ser (spec ...)")
	}

	spec := &lexerSpec{languages: make(map[string][]tokenDef)}
	for _, sectionValue := range root[1:] {
		section, ok := sectionValue.([]any)
		if !ok || len(section) == 0 {
			continue
		}

		head, ok := section[0].(string)
		if !ok || head != "language" {
			continue
		}
		if len(section) < 2 {
			return nil, errors.New("seccion language invalida")
		}

		language, ok := section[1].(string)
		if !ok {
			return nil, errors.New("nombre de lenguaje invalido")
		}

		var defs []tokenDef
		for _, itemValue := range section[2:] {
			item, ok := itemValue.([]any)
			if !ok || len(item) < 3 {
				continue
			}

			kind, ok := item[0].(string)
			if !ok || (kind != "token" && kind != "skip") {
				continue
			}
			name, ok := item[1].(string)
			if !ok {
				return nil, errors.New("nombre de token invalido")
			}

			pattern, err := compileRegex(item[2])
			if err != nil {
				return nil, fmt.Errorf("%s.%s: %w", language, name, err)
			}
			compiled, err := regexp.Compile("^(?:" + pattern + ")")
			if err != nil {
				return nil, fmt.Errorf("%s.%s: regex invalida: %w", language, name, err)
			}
			compiled.Longest()

			defs = append(defs, tokenDef{
				name:    name,
				pattern: pattern,
				regex:   compiled,
				skip:    kind == "skip",
			})
		}

		spec.languages[language] = defs
	}

	return spec, nil
}

func compileRegex(expr any) (string, error) {
	if symbol, ok := expr.(string); ok {
		if symbol == "dot" {
			return "(?s:.)", nil
		}
		return "", fmt.Errorf("regex invalida: simbolo suelto %q", symbol)
	}

	form, ok := expr.([]any)
	if !ok || len(form) == 0 {
		return "", errors.New("regex vacia")
	}

	op, ok := form[0].(string)
	if !ok {
		return "", errors.New("operador regex invalido")
	}
	args := form[1:]

	switch op {
	case "lit":
		if len(args) != 1 {
			return "", errors.New("lit requiere un string")
		}
		value, ok := args[0].(string)
		if !ok {
			return "", errors.New("lit requiere un string")
		}
		return regexp.QuoteMeta(value), nil
	case "charset":
		if len(args) != 1 {
			return "", errors.New("charset requiere un string")
		}
		value, ok := args[0].(string)
		if !ok {
			return "", errors.New("charset requiere un string")
		}
		return compileCharClass(value, false), nil
	case "ncharset":
		if len(args) != 1 {
			return "", errors.New("ncharset requiere un string")
		}
		value, ok := args[0].(string)
		if !ok {
			return "", errors.New("ncharset requiere un string")
		}
		return compileCharClass(value, true), nil
	case "dot":
		if len(args) != 0 {
			return "", errors.New("dot no recibe argumentos")
		}
		return "(?s:.)", nil
	case "seq":
		if len(args) == 0 {
			return "", errors.New("seq requiere al menos un argumento")
		}
		var out strings.Builder
		for _, arg := range args {
			child, err := compileRegex(arg)
			if err != nil {
				return "", err
			}
			out.WriteString(child)
		}
		return out.String(), nil
	case "or":
		if len(args) == 0 {
			return "", errors.New("or requiere al menos un argumento")
		}
		parts := make([]string, 0, len(args))
		for _, arg := range args {
			child, err := compileRegex(arg)
			if err != nil {
				return "", err
			}
			parts = append(parts, child)
		}
		return "(?:" + strings.Join(parts, "|") + ")", nil
	case "star", "plus", "opt":
		if len(args) != 1 {
			return "", fmt.Errorf("%s requiere un argumento", op)
		}
		child, err := compileRegex(args[0])
		if err != nil {
			return "", err
		}
		switch op {
		case "star":
			return "(?:" + child + ")*", nil
		case "plus":
			return "(?:" + child + ")+", nil
		default:
			return "(?:" + child + ")?", nil
		}
	default:
		return "", fmt.Errorf("operador regex desconocido: %s", op)
	}
}

func compileCharClass(pattern string, negated bool) string {
	chars := expandCharset(pattern)
	seen := make(map[rune]bool, len(chars))
	unique := make([]rune, 0, len(chars))
	for _, ch := range chars {
		if !seen[ch] {
			seen[ch] = true
			unique = append(unique, ch)
		}
	}
	sort.Slice(unique, func(i, j int) bool { return unique[i] < unique[j] })

	var out strings.Builder
	out.WriteByte('[')
	if negated {
		out.WriteByte('^')
	}
	for _, ch := range unique {
		out.WriteString(escapeCharClassRune(ch))
	}
	out.WriteByte(']')
	return out.String()
}

func expandCharset(pattern string) []rune {
	runes := []rune(pattern)
	chars := make([]rune, 0, len(runes))

	for i := 0; i < len(runes); {
		if i+2 < len(runes) && runes[i+1] == '-' {
			start := runes[i]
			end := runes[i+2]
			if start <= end {
				for ch := start; ch <= end; ch++ {
					chars = append(chars, ch)
				}
			} else {
				for ch := end; ch <= start; ch++ {
					chars = append(chars, ch)
				}
			}
			i += 3
			continue
		}

		chars = append(chars, runes[i])
		i++
	}

	return chars
}

func escapeCharClassRune(ch rune) string {
	switch ch {
	case '\\':
		return `\\`
	case ']':
		return `\]`
	case '[':
		return `\[`
	case '^':
		return `\^`
	case '-':
		return `\-`
	case '\n':
		return `\n`
	case '\r':
		return `\r`
	case '\t':
		return `\t`
	case 0:
		return `\x00`
	default:
		if ch < 32 || ch == 127 {
			return fmt.Sprintf(`\x%02x`, ch)
		}
		return string(ch)
	}
}

func scanText(text string, defs []tokenDef) []token {
	var tokens []token
	pos := 0
	line := 1
	col := 1

	for pos < len(text) {
		bestLen := 0
		bestIndex := -1
		remaining := text[pos:]

		for i, def := range defs {
			match := def.regex.FindStringIndex(remaining)
			if match == nil || match[0] != 0 {
				continue
			}
			size := match[1]
			if size > bestLen {
				bestLen = size
				bestIndex = i
			}
		}

		if bestLen == 0 || bestIndex == -1 {
			_, width := utf8.DecodeRuneInString(remaining)
			if width == 0 {
				break
			}
			lexeme := remaining[:width]
			tokens = append(tokens, token{typ: "ERROR", lexeme: lexeme, line: line, col: col})
			line, col = advancePosition(lexeme, line, col)
			pos += width
			continue
		}

		lexeme := text[pos : pos+bestLen]
		def := defs[bestIndex]
		if !def.skip {
			tokens = append(tokens, token{typ: def.name, lexeme: lexeme, line: line, col: col})
		}

		line, col = advancePosition(lexeme, line, col)
		pos += bestLen
	}

	return tokens
}

func advancePosition(text string, line int, col int) (int, int) {
	for _, ch := range text {
		if ch == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

func formatTokens(tokens []token) string {
	lines := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		lines = append(lines, fmt.Sprintf("%-14s | L%3d:C%-3d | %s", tok.typ, tok.line, tok.col, pythonRepr(tok.lexeme)))
	}
	return strings.Join(lines, "\n")
}

func pythonRepr(text string) string {
	var out strings.Builder
	out.WriteByte('\'')
	for _, ch := range text {
		switch ch {
		case '\\':
			out.WriteString(`\\`)
		case '\'':
			out.WriteString(`\'`)
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		case 0:
			out.WriteString(`\x00`)
		default:
			if ch < 32 || ch == 127 {
				out.WriteString(fmt.Sprintf(`\x%02x`, ch))
			} else {
				out.WriteRune(ch)
			}
		}
	}
	out.WriteByte('\'')
	return out.String()
}

func resolveLanguage(requested string, path string, spec *lexerSpec) (string, error) {
	if requested != "auto" {
		if _, ok := spec.languages[requested]; !ok {
			return "", fmt.Errorf("lenguaje no definido en spec: %s", requested)
		}
		return requested, nil
	}

	language, ok := languageByExtension[strings.ToLower(filepath.Ext(path))]
	if !ok {
		return "", fmt.Errorf("no se pudo detectar el lenguaje de: %s", path)
	}
	if _, ok := spec.languages[language]; !ok {
		return "", fmt.Errorf("lenguaje no definido en spec: %s", language)
	}
	return language, nil
}

func scanFile(index int, path string, requestedLanguage string, spec *lexerSpec) fileResult {
	start := time.Now()

	language, err := resolveLanguage(requestedLanguage, path, spec)
	if err != nil {
		return fileResult{index: index, input: path, err: err}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fileResult{index: index, input: path, language: language, err: err}
	}

	tokens := scanText(string(content), spec.languages[language])
	return fileResult{
		index:    index,
		input:    path,
		language: language,
		tokens:   tokens,
		duration: time.Since(start),
	}
}

func processFiles(paths []string, requestedLanguage string, spec *lexerSpec, workers int) []fileResult {
	results := make([]fileResult, len(paths))
	resultCh := make(chan fileResult, len(paths))
	sem := make(chan struct{}, workers)

	var wg sync.WaitGroup
	for index, path := range paths {
		wg.Add(1)
		go func(index int, path string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			resultCh <- scanFile(index, path, requestedLanguage, spec)
		}(index, path)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for result := range resultCh {
		results[result.index] = result
	}
	return results
}

func renderResult(result fileResult, includeHeader bool) string {
	body := formatTokens(result.tokens)
	if !includeHeader {
		return body
	}
	return fmt.Sprintf("== %s (%s) ==\n%s", result.input, result.language, body)
}

func writeOutputs(results []fileResult, outputPath string, outputDir string) error {
	var problems []string
	for _, result := range results {
		if result.err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", result.input, result.err))
		}
	}
	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "\n"))
	}

	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return err
		}
		used := make(map[string]int)
		for _, result := range results {
			path := uniqueOutputPath(outputDir, result.input, used)
			if err := os.WriteFile(path, []byte(renderResult(result, false)+"\n"), 0o644); err != nil {
				return err
			}
		}
		return nil
	}

	parts := make([]string, 0, len(results))
	includeHeader := len(results) > 1
	for _, result := range results {
		parts = append(parts, renderResult(result, includeHeader))
	}
	out := strings.Join(parts, "\n\n")

	if outputPath != "" {
		return os.WriteFile(outputPath, []byte(out+"\n"), 0o644)
	}

	fmt.Println(out)
	return nil
}

func uniqueOutputPath(outputDir string, input string, used map[string]int) string {
	base := filepath.Base(input)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	if stem == "" {
		stem = "tokens"
	}

	path := filepath.Join(outputDir, stem+"_tokens.txt")
	count := used[path]
	used[path] = count + 1
	if count == 0 {
		return path
	}
	return filepath.Join(outputDir, fmt.Sprintf("%s_%d_tokens.txt", stem, count+1))
}

func run(args []string) error {
	var inputs inputList
	fs := flag.NewFlagSet("motordelexico", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	specPath := fs.String("spec", "lexical_spec.sexp", "Ruta al archivo de especificacion .sexp")
	language := fs.String("language", "go", "Lenguaje a escanear: erlang, go, cpp o auto")
	outputPath := fs.String("output", "", "Archivo de salida opcional; con varios inputs escribe una salida combinada")
	outputDir := fs.String("output-dir", "", "Directorio opcional para escribir un archivo de tokens por input")
	workers := fs.Int("workers", 4, "Cantidad maxima de archivos procesados concurrentemente")
	showTime := fs.Bool("time", false, "Muestra el tiempo total en stderr")
	fs.Var(&inputs, "input", "Archivo fuente a tokenizar; puede repetirse")

	if err := fs.Parse(args); err != nil {
		return err
	}
	inputs = append(inputs, fs.Args()...)

	if len(inputs) == 0 {
		return errors.New("se requiere al menos un --input o argumento posicional")
	}
	if *outputPath != "" && *outputDir != "" {
		return errors.New("--output y --output-dir no se pueden usar al mismo tiempo")
	}
	if *workers < 1 {
		return errors.New("--workers debe ser mayor o igual que 1")
	}

	specBytes, err := os.ReadFile(*specPath)
	if err != nil {
		return err
	}
	spec, err := parseSpec(string(specBytes))
	if err != nil {
		return err
	}

	start := time.Now()
	results := processFiles(inputs, *language, spec, *workers)
	if err := writeOutputs(results, *outputPath, *outputDir); err != nil {
		return err
	}
	if *showTime {
		fmt.Fprintf(os.Stderr, "Tiempo total: %s\n", time.Since(start))
	}

	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
