package main

import (
	"os"
	"strings"
	"testing"
)

func loadTestSpec(t *testing.T) *lexerSpec {
	t.Helper()

	specBytes, err := os.ReadFile("lexical_spec.sexp")
	if err != nil {
		t.Fatal(err)
	}
	spec, err := parseSpec(string(specBytes))
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func TestScanGoExampleMatchesPythonOutput(t *testing.T) {
	spec := loadTestSpec(t)

	source, err := os.ReadFile("example.go")
	if err != nil {
		t.Fatal(err)
	}
	expected, err := os.ReadFile("tokens_go.txt")
	if err != nil {
		t.Fatal(err)
	}

	tokens := scanText(string(source), spec.languages["go"])
	got := formatTokens(tokens)
	want := strings.TrimSuffix(string(expected), "\n")

	if got != want {
		t.Fatalf("salida distinta\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestConcurrentAutoLanguageKeepsInputOrder(t *testing.T) {
	spec := loadTestSpec(t)

	inputs := []string{"example.go", "example.cpp", "example.erl"}
	results := processFiles(inputs, "auto", spec, 3)

	for index, result := range results {
		if result.err != nil {
			t.Fatalf("resultado %d fallo: %v", index, result.err)
		}
		if result.input != inputs[index] {
			t.Fatalf("orden incorrecto en %d: got %q want %q", index, result.input, inputs[index])
		}
	}

	if results[0].language != "go" || results[1].language != "cpp" || results[2].language != "erlang" {
		t.Fatalf("deteccion auto incorrecta: %#v", []string{results[0].language, results[1].language, results[2].language})
	}
}

func TestTokenRegexUsesLongestAlternative(t *testing.T) {
	spec := loadTestSpec(t)

	erlangTokens := scanText("andalso", spec.languages["erlang"])
	if len(erlangTokens) != 1 || erlangTokens[0].typ != "KEYWORD" || erlangTokens[0].lexeme != "andalso" {
		t.Fatalf("andalso no fue reconocido como keyword completa: %#v", erlangTokens)
	}

	cppTokens := scanText("char8_t", spec.languages["cpp"])
	if len(cppTokens) != 1 || cppTokens[0].typ != "KEYWORD" || cppTokens[0].lexeme != "char8_t" {
		t.Fatalf("char8_t no fue reconocido como keyword completa: %#v", cppTokens)
	}
}
