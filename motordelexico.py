#!/usr/bin/env python3

from __future__ import annotations

import argparse
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, Iterable, List, Sequence, Set, Tuple

RegexNode = Tuple
TokenDef = Tuple[str, RegexNode, bool]  # (nombre, patron, skip)


class SexpParser:
    """Parseador simple de expresiones-s con soporte de strings y comentarios ';'."""

    def __init__(self, text: str) -> None:
        self.text = text
        self.i = 0
        self.n = len(text)

    def parse(self) -> List:
        items: List = []
        while True:
            self._skip_ws_and_comments()
            if self.i >= self.n:
                break
            items.append(self._parse_expr())
        return items

    def _skip_ws_and_comments(self) -> None:
        while self.i < self.n:
            ch = self.text[self.i]
            if ch in " \t\r\n":
                self.i += 1
                continue
            if ch == ";":
                while self.i < self.n and self.text[self.i] != "\n":
                    self.i += 1
                continue
            break

    def _parse_expr(self):
        self._skip_ws_and_comments()
        if self.i >= self.n:
            raise ValueError("Fin inesperado al parsear expresion-s")

        ch = self.text[self.i]
        if ch == "(":
            self.i += 1
            out = []
            while True:
                self._skip_ws_and_comments()
                if self.i >= self.n:
                    raise ValueError("Parentesis sin cerrar en expresion-s")
                if self.text[self.i] == ")":
                    self.i += 1
                    return out
                out.append(self._parse_expr())
        if ch == '"':
            return self._parse_string()
        if ch == ")":
            raise ValueError("Parentesis de cierre inesperado")
        return self._parse_symbol()

    def _parse_string(self) -> str:
        assert self.text[self.i] == '"'
        self.i += 1
        out: List[str] = []

        escapes = {
            "n": "\n",
            "r": "\r",
            "t": "\t",
            "\\": "\\",
            '"': '"',
            "'": "'",
            "0": "\0",
        }

        while self.i < self.n:
            ch = self.text[self.i]
            self.i += 1
            if ch == '"':
                return "".join(out)
            if ch == "\\":
                if self.i >= self.n:
                    raise ValueError("Escape incompleto en string")
                esc = self.text[self.i]
                self.i += 1
                out.append(escapes.get(esc, esc))
                continue
            out.append(ch)

        raise ValueError("String sin cerrar")

    def _parse_symbol(self) -> str:
        start = self.i
        while self.i < self.n and self.text[self.i] not in "() \t\r\n":
            self.i += 1
        if start == self.i:
            raise ValueError("Simbolo vacio")
        return self.text[start : self.i]


@dataclass
class LexerSpec:
    common_categories: List[Tuple[str, str]]
    languages: Dict[str, List[TokenDef]]


def parse_charset(pattern: str) -> Set[str]:
    """Convierte un patron tipo 'A-Za-z0-9_' en conjunto de caracteres."""
    chars: Set[str] = set()
    i = 0
    n = len(pattern)

    while i < n:
        ch = pattern[i]
        if i + 2 < n and pattern[i + 1] == "-":
            start = ord(ch)
            end = ord(pattern[i + 2])
            if start <= end:
                for code in range(start, end + 1):
                    chars.add(chr(code))
            else:
                for code in range(end, start + 1):
                    chars.add(chr(code))
            i += 3
            continue

        chars.add(ch)
        i += 1

    return chars


def compile_regex(expr) -> RegexNode:
    if isinstance(expr, str):
        if expr == "dot":
            return ("dot",)
        raise ValueError(f"Regex invalida: simbolo suelto '{expr}'")
    if not expr:
        raise ValueError("Regex vacia")

    op = expr[0]
    args = expr[1:]

    if op == "lit":
        if len(args) != 1 or not isinstance(args[0], str):
            raise ValueError("lit requiere un string")
        return ("lit", args[0])

    if op == "charset":
        if len(args) != 1 or not isinstance(args[0], str):
            raise ValueError("charset requiere un string")
        return ("charset", frozenset(parse_charset(args[0])))

    if op == "ncharset":
        if len(args) != 1 or not isinstance(args[0], str):
            raise ValueError("ncharset requiere un string")
        return ("ncharset", frozenset(parse_charset(args[0])))

    if op == "dot":
        if args:
            raise ValueError("dot no recibe argumentos")
        return ("dot",)

    if op in {"seq", "or"}:
        if not args:
            raise ValueError(f"{op} requiere al menos un argumento")
        return (op, tuple(compile_regex(a) for a in args))

    if op in {"star", "plus", "opt"}:
        if len(args) != 1:
            raise ValueError(f"{op} requiere un argumento")
        return (op, compile_regex(args[0]))

    raise ValueError(f"Operador regex desconocido: {op}")


def _match(node: RegexNode, text: str, pos: int, memo: Dict[Tuple[RegexNode, int], Set[int]]) -> Set[int]:
    key = (node, pos)
    if key in memo:
        return memo[key]

    op = node[0]
    result: Set[int]

    if op == "lit":
        lit = node[1]
        if text.startswith(lit, pos):
            result = {pos + len(lit)}
        else:
            result = set()

    elif op == "charset":
        allowed = node[1]
        if pos < len(text) and text[pos] in allowed:
            result = {pos + 1}
        else:
            result = set()

    elif op == "ncharset":
        banned = node[1]
        if pos < len(text) and text[pos] not in banned:
            result = {pos + 1}
        else:
            result = set()

    elif op == "dot":
        result = {pos + 1} if pos < len(text) else set()

    elif op == "or":
        result = set()
        for child in node[1]:
            result |= _match(child, text, pos, memo)

    elif op == "seq":
        positions = {pos}
        for child in node[1]:
            nxt: Set[int] = set()
            for p in positions:
                nxt |= _match(child, text, p, memo)
            positions = nxt
            if not positions:
                break
        result = positions

    elif op == "opt":
        result = {pos} | _match(node[1], text, pos, memo)

    elif op == "star":
        # Cierre transitivo sobre avances de la subexpresion
        child = node[1]
        visited = {pos}
        stack = [pos]
        while stack:
            p = stack.pop()
            for q in _match(child, text, p, memo):
                if q not in visited:
                    visited.add(q)
                    stack.append(q)
        result = visited

    elif op == "plus":
        child = node[1]
        result = set()
        for mid in _match(child, text, pos, memo):
            result |= _match(("star", child), text, mid, memo)

    else:
        raise ValueError(f"Nodo regex no soportado: {op}")

    memo[key] = result
    return result


def match_prefix(node: RegexNode, text: str, pos: int) -> int:
    """Devuelve la longitud maxima que matchea desde pos. 0 si no hay match."""
    memo: Dict[Tuple[RegexNode, int], Set[int]] = {}
    ends = _match(node, text, pos, memo)
    if not ends:
        return 0
    return max(ends) - pos


def parse_spec(spec_text: str) -> LexerSpec:
    parsed = SexpParser(spec_text).parse()
    if len(parsed) != 1:
        raise ValueError("El archivo de especificacion debe tener una unica forma raiz")

    root = parsed[0]
    if not isinstance(root, list) or not root or root[0] != "spec":
        raise ValueError("La raiz debe ser (spec ...)")

    common_categories: List[Tuple[str, str]] = []
    languages: Dict[str, List[TokenDef]] = {}

    for section in root[1:]:
        if not isinstance(section, list) or not section:
            continue

        head = section[0]

        if head == "common-categories":
            for entry in section[1:]:
                if isinstance(entry, list) and len(entry) == 2 and isinstance(entry[0], str) and isinstance(entry[1], str):
                    common_categories.append((entry[0], entry[1]))

        elif head == "language":
            if len(section) < 2 or not isinstance(section[1], str):
                raise ValueError("Seccion language invalida")
            lang = section[1]
            token_defs: List[TokenDef] = []

            for item in section[2:]:
                if not isinstance(item, list) or len(item) < 3:
                    continue
                kind = item[0]
                token_name = item[1]
                regex_form = item[2]

                if kind not in {"token", "skip"}:
                    continue
                if not isinstance(token_name, str):
                    raise ValueError("Nombre de token invalido")

                compiled = compile_regex(regex_form)
                token_defs.append((token_name, compiled, kind == "skip"))

            languages[lang] = token_defs

    return LexerSpec(common_categories=common_categories, languages=languages)


def scan_text(text: str, token_defs: Sequence[TokenDef]) -> List[Tuple[str, str, int, int]]:
    """Escanea texto y devuelve tokens como (tipo, lexema, linea, columna)."""
    tokens: List[Tuple[str, str, int, int]] = []
    pos = 0
    line = 1
    col = 1

    while pos < len(text):
        best_len = 0
        best_token: TokenDef | None = None

        for token_def in token_defs:
            token_name, regex_node, _skip = token_def
            size = match_prefix(regex_node, text, pos)
            if size > best_len:
                best_len = size
                best_token = token_def
            elif size == best_len and size > 0 and best_token is not None:
                # desempate por prioridad (orden en el archivo): conservar el primero
                _ = token_name

        if best_len == 0 or best_token is None:
            bad_char = text[pos]
            tokens.append(("ERROR", bad_char, line, col))
            pos += 1
            if bad_char == "\n":
                line += 1
                col = 1
            else:
                col += 1
            continue

        token_name, _regex_node, skip = best_token
        lexeme = text[pos : pos + best_len]

        if not skip:
            tokens.append((token_name, lexeme, line, col))

        for ch in lexeme:
            if ch == "\n":
                line += 1
                col = 1
            else:
                col += 1

        pos += best_len

    return tokens


def cli() -> int:
    parser = argparse.ArgumentParser(description="Motor regex + scanner lexico desde especificacion s-exp")
    parser.add_argument("--spec", required=True, help="Ruta al archivo de especificacion .sexp")
    parser.add_argument("--language", required=True, choices=["erlang", "go", "cpp"], help="Lenguaje a escanear")
    parser.add_argument("--input", required=True, help="Archivo fuente a tokenizar")
    parser.add_argument("--output", help="Archivo de salida opcional")

    args = parser.parse_args()

    spec_path = Path(args.spec)
    src_path = Path(args.input)

    spec = parse_spec(spec_path.read_text(encoding="utf-8"))
    if args.language not in spec.languages:
        raise ValueError(f"Lenguaje no definido en spec: {args.language}")

    text = src_path.read_text(encoding="utf-8")
    token_defs = spec.languages[args.language]
    tokens = scan_text(text, token_defs)

    lines = [f"{ttype:<14} | L{ln:>3}:C{col:<3} | {lex!r}" for ttype, lex, ln, col in tokens]
    out = "\n".join(lines)

    if args.output:
        Path(args.output).write_text(out + "\n", encoding="utf-8")
    else:
        print(out)

    return 0


if __name__ == "__main__":
    raise SystemExit(cli())
