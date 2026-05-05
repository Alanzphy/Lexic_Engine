(spec
  (title "Especificacion lexica para Erlang, Go y C++")

  (common-categories
    (KEYWORD "Palabras reservadas")
    (IDENTIFIER "Nombres de entidades")
    (LITERALS "Numericos, string, char, booleanos")
    (OPERATORS "Operadores aritmeticos, logicos, asignacion, comparacion")
    (PUNCTUATION "Separadores y delimitadores")
    (COMMENT "Comentarios")
    (WHITESPACE "Espacios, tabs y saltos de linea"))

  (notation
    (lit "cadena literal exacta")
    (charset "conjunto de caracteres permitidos, soporta rangos como A-Z")
    (ncharset "negacion de charset")
    (seq "concatenacion")
    (or "alternacion")
    (star "repeticion 0..n")
    (plus "repeticion 1..n")
    (opt "opcional 0..1")
    (dot "cualquier caracter"))

  (language erlang
    (skip WHITESPACE (plus (charset " \t\r\n")))

    (token COMMENT
      (seq (lit "%") (star (ncharset "\n"))))

    (token KEYWORD
      (or
        (lit "after") (lit "and") (lit "andalso") (lit "band") (lit "begin")
        (lit "bnot") (lit "bor") (lit "bsl") (lit "bsr") (lit "bxor")
        (lit "case") (lit "catch") (lit "cond") (lit "div") (lit "else")
        (lit "end") (lit "fun") (lit "if") (lit "let") (lit "maybe")
        (lit "not") (lit "of") (lit "or") (lit "orelse") (lit "receive")
        (lit "rem") (lit "try") (lit "when") (lit "xor")))

    (token STRING
      (seq (lit "\"")
           (star (or (seq (lit "\\") dot)
                     (ncharset "\"\n")))
           (lit "\"")))

    (token QUOTED_ATOM
      (seq (lit "'")
           (star (or (seq (lit "\\") dot)
                     (ncharset "'\n")))
           (lit "'")))

    (token FLOAT
      (seq (plus (charset "0-9"))
           (lit ".")
           (plus (charset "0-9"))
           (opt (seq (or (lit "e") (lit "E"))
                     (opt (or (lit "+") (lit "-")))
                     (plus (charset "0-9"))))))

    (token BASE_INT
      (seq (plus (charset "0-9"))
           (lit "#")
           (plus (charset "0-9A-Za-z"))))

    (token INTEGER
      (plus (charset "0-9")))

    (token VARIABLE
      (seq (charset "A-Z_")
           (star (charset "A-Za-z0-9_@"))))

    (token ATOM
      (seq (charset "a-z")
           (star (charset "A-Za-z0-9_@"))))

    (token OPERATOR
      (or
        (lit "=:=") (lit "=/=") (lit "==") (lit "/=")
        (lit "=<") (lit ">=") (lit "<") (lit ">")
        (lit "=") (lit "!")
        (lit "++") (lit "--")
        (lit "+") (lit "-") (lit "*") (lit "/")))

    (token PUNCTUATION
      (or
        (lit "->") (lit "=>")
        (lit "(") (lit ")") (lit "[") (lit "]") (lit "{") (lit "}")
        (lit ",") (lit ";") (lit ".") (lit ":"))))

  (language go
    (skip WHITESPACE (plus (charset " \t\r\n")))

    (token COMMENT
      (or
        (seq (lit "//") (star (ncharset "\n")))
        (seq (lit "/*")
             (star (or (ncharset "*")
                       (seq (lit "*") (ncharset "/"))))
             (lit "*/"))))

    (token KEYWORD
      (or
        (lit "break") (lit "default") (lit "func") (lit "interface") (lit "select")
        (lit "case") (lit "defer") (lit "go") (lit "map") (lit "struct")
        (lit "chan") (lit "else") (lit "goto") (lit "package") (lit "switch")
        (lit "const") (lit "fallthrough") (lit "if") (lit "range") (lit "type")
        (lit "continue") (lit "for") (lit "import") (lit "return") (lit "var")))

    (token BOOL_LITERAL (or (lit "true") (lit "false")))
    (token NIL_LITERAL (lit "nil"))

    (token RAW_STRING
      (seq (lit "`") (star (ncharset "`")) (lit "`")))

    (token STRING
      (seq (lit "\"")
           (star (or (seq (lit "\\") dot)
                     (ncharset "\"\n")))
           (lit "\"")))

    (token RUNE
      (seq (lit "'")
           (or (seq (lit "\\") dot)
               (ncharset "'\n"))
           (lit "'")))

    (token IMAG_NUMBER
      (seq
        (or
          (seq (plus (charset "0-9")) (lit ".") (star (charset "0-9")))
          (seq (lit ".") (plus (charset "0-9")))
          (plus (charset "0-9")))
        (lit "i")))

    (token FLOAT
      (seq
        (or
          (seq (plus (charset "0-9")) (lit ".") (star (charset "0-9")))
          (seq (lit ".") (plus (charset "0-9"))))
        (opt (seq (or (lit "e") (lit "E"))
                  (opt (or (lit "+") (lit "-")))
                  (plus (charset "0-9"))))))

    (token INTEGER
      (or
        (seq (lit "0") (or (lit "x") (lit "X")) (plus (charset "0-9A-Fa-f_")))
        (seq (lit "0") (or (lit "b") (lit "B")) (plus (charset "01_")))
        (seq (lit "0") (or (lit "o") (lit "O")) (plus (charset "0-7_")))
        (plus (charset "0-9_"))))

    (token IDENTIFIER
      (seq (charset "A-Za-z_")
           (star (charset "A-Za-z0-9_"))))

    (token OPERATOR
      (or
        (lit "<<=") (lit ">>=") (lit "&^=")
        (lit "++") (lit "--") (lit "&&") (lit "||")
        (lit "==") (lit "!=") (lit "<=") (lit ">=") (lit ":=")
        (lit "+=") (lit "-=") (lit "*=") (lit "/=") (lit "%=")
        (lit "&=") (lit "|=") (lit "^=")
        (lit "<<") (lit ">>") (lit "&^")
        (lit "<-")
        (lit "+") (lit "-") (lit "*") (lit "/") (lit "%")
        (lit "&") (lit "|") (lit "^") (lit "!")
        (lit "=") (lit "<") (lit ">") (lit "~")))

    (token PUNCTUATION
      (or
        (lit "...")
        (lit "(") (lit ")") (lit "[") (lit "]") (lit "{") (lit "}")
        (lit ",") (lit ";") (lit ".") (lit ":"))))

  (language cpp
    (skip WHITESPACE (plus (charset " \t\r\n")))

    (token PREPROCESSOR
      (seq (lit "#") (star (ncharset "\n"))))

    (token COMMENT
      (or
        (seq (lit "//") (star (ncharset "\n")))
        (seq (lit "/*")
             (star (or (ncharset "*")
                       (seq (lit "*") (ncharset "/"))))
             (lit "*/"))))

    (token KEYWORD
      (or
        (lit "alignas") (lit "alignof") (lit "and") (lit "and_eq") (lit "asm")
        (lit "auto") (lit "bitand") (lit "bitor") (lit "bool") (lit "break")
        (lit "case") (lit "catch") (lit "char") (lit "char8_t") (lit "char16_t")
        (lit "char32_t") (lit "class") (lit "compl") (lit "concept") (lit "const")
        (lit "consteval") (lit "constexpr") (lit "constinit") (lit "const_cast")
        (lit "continue") (lit "co_await") (lit "co_return") (lit "co_yield")
        (lit "decltype") (lit "default") (lit "delete") (lit "do") (lit "double")
        (lit "dynamic_cast") (lit "else") (lit "enum") (lit "explicit") (lit "export")
        (lit "extern") (lit "false") (lit "float") (lit "for") (lit "friend")
        (lit "goto") (lit "if") (lit "inline") (lit "int") (lit "long")
        (lit "mutable") (lit "namespace") (lit "new") (lit "noexcept")
        (lit "not") (lit "not_eq") (lit "nullptr") (lit "operator") (lit "or")
        (lit "or_eq") (lit "private") (lit "protected") (lit "public")
        (lit "register") (lit "reinterpret_cast") (lit "requires") (lit "return")
        (lit "short") (lit "signed") (lit "sizeof") (lit "static")
        (lit "static_assert") (lit "static_cast") (lit "struct") (lit "switch")
        (lit "template") (lit "this") (lit "thread_local") (lit "throw")
        (lit "true") (lit "try") (lit "typedef") (lit "typeid") (lit "typename")
        (lit "union") (lit "unsigned") (lit "using") (lit "virtual") (lit "void")
        (lit "volatile") (lit "wchar_t") (lit "while") (lit "xor") (lit "xor_eq")))

    (token STRING
      (seq
        (opt (or (lit "u8") (lit "u") (lit "U") (lit "L")))
        (lit "\"")
        (star (or (seq (lit "\\") dot)
                  (ncharset "\"\n")))
        (lit "\"")))

    (token CHAR
      (seq
        (opt (or (lit "u8") (lit "u") (lit "U") (lit "L")))
        (lit "'")
        (or (seq (lit "\\") dot)
            (ncharset "'\n"))
        (lit "'")))

    (token NUMBER
      (or
        (seq (plus (charset "0-9")) (lit ".") (star (charset "0-9"))
             (opt (seq (or (lit "e") (lit "E"))
                       (opt (or (lit "+") (lit "-")))
                       (plus (charset "0-9")))))
        (seq (lit ".") (plus (charset "0-9"))
             (opt (seq (or (lit "e") (lit "E"))
                       (opt (or (lit "+") (lit "-")))
                       (plus (charset "0-9")))))
        (seq (lit "0") (or (lit "x") (lit "X")) (plus (charset "0-9A-Fa-f")))
        (plus (charset "0-9"))))

    (token IDENTIFIER
      (seq (charset "A-Za-z_")
           (star (charset "A-Za-z0-9_"))))

    (token OPERATOR
      (or
        (lit "<<=") (lit ">>=")
        (lit "->*") (lit "->") (lit "::")
        (lit "++") (lit "--")
        (lit "&&") (lit "||")
        (lit "==") (lit "!=") (lit "<=") (lit ">=")
        (lit "+=") (lit "-=") (lit "*=") (lit "/=") (lit "%=")
        (lit "&=") (lit "|=") (lit "^=")
        (lit "<<") (lit ">>")
        (lit ".*")
        (lit "=") (lit "+") (lit "-") (lit "*") (lit "/") (lit "%")
        (lit "&") (lit "|") (lit "^") (lit "~")
        (lit "!") (lit "<") (lit ">") (lit "?")))

    (token PUNCTUATION
      (or
        (lit "(") (lit ")") (lit "[") (lit "]") (lit "{") (lit "}")
        (lit ",") (lit ";") (lit ".") (lit ":"))))
)
