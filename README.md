# Analizador Lexico con Expresiones-S

## 1. Lenguajes seleccionados y categorias lexicas comunes

Lenguajes seleccionados:
- **Erlang** (paradigma funcional)
- **Go** (paradigma imperativo/procedimental)
- **C++** (paradigma orientado a objetos)

Categorias lexicas que comparten (nivel general):
- **Palabras reservadas**
- **Identificadores**
- **Literales** (numericos, string, char/bool segun el lenguaje)
- **Operadores**
- **Separadores y delimitadores**
- **Comentarios**
- **Espacios en blanco**

> Diferencias importantes: cada lenguaje cambia en los detalles (por ejemplo, comentarios `%` en Erlang y `//`/`/* */` en Go/C++, o variables de Erlang que inician con mayuscula/`_`).

## 2. Notacion basada en expresiones-s para regex

Se definio una notacion s-exp para describir regex de forma estructurada:

- `(lit "texto")`: literal exacto
- `(charset "A-Za-z0-9_")`: conjunto/rangos de caracteres
- `(ncharset "\n")`: negacion de conjunto
- `(seq r1 r2 ...)`: concatenacion
- `(or r1 r2 ...)`: alternacion
- `(star r)`: repeticion 0..n
- `(plus r)`: repeticion 1..n
- `(opt r)`: opcional 0..1
- `dot`: cualquier caracter

## 3. Representacion de categorias lexicas por lenguaje

La representacion completa se encuentra en el archivo:
- `lexical_spec.sexp`

Estructura del archivo:
- `common-categories`: categorias compartidas
- `notation`: definicion formal de la notacion s-exp
- `language erlang|go|cpp`: tokens concretos de cada lenguaje
- `token` vs `skip`: tokens reportables y tokens ignorables (`WHITESPACE`)

## 4. Implementacion del motor regex y scanner

Implementacion en Python:
- `motordelexico.py`

### Arquitectura
1. **Parser de s-exp**
   - Lee `lexical_spec.sexp`.
   - Construye una estructura interna para cada token.

2. **Compilacion regex**
   - Convierte cada s-exp a un AST interno (`lit`, `seq`, `or`, `star`, etc.).

3. **Motor de matching**
   - Evalua cada AST con un algoritmo recursivo + memoizacion.
   - `star` usa cierre transitivo para evitar loops infinitos.

4. **Scanner lexico**
   - Recorre el archivo fuente posicion por posicion.
   - Aplica *longest match* (elige el token de mayor longitud).
   - Si hay empate, respeta prioridad por orden en la especificacion.
   - Reporta `ERROR` cuando ningun token coincide.

## 5. Convenciones de codificacion usadas

Se siguieron convenciones de Python:
- Nombres `snake_case`
- Tipado con `typing`
- `dataclass` para estructura de especificacion
- Funciones con responsabilidad clara y documentadas de forma breve
- CLI con `argparse`

## 6. Reflexion tecnica

### Solucion planteada
Separar **especificacion** (archivo `.sexp`) de **ejecucion** (motor) permite:
- modificar reglas lexicas sin tocar el motor,
- reutilizar el mismo escaner para Erlang, Go y C++,
- priorizar trazabilidad para la entrega.

### Algoritmos implementados
- Parser descendente para expresiones-s.
- Evaluacion de regex por arbol de sintaxis abstracto.
- Memoizacion por estado `(nodo_regex, posicion)` para evitar recomputos.
- Cierre para `star` mediante exploracion de estados alcanzables.
- Seleccion de token por *longest match*.

### Complejidad aproximada
Sea:
- `N` = longitud del archivo fuente,
- `T` = cantidad de tokens definidos para un lenguaje,
- `M` = costo promedio de evaluar un token en una posicion.

Costo de escaneo: **O(N * T * M)**.

En esta implementacion, `M` se reduce con memoizacion local por intento de match.
En entradas reales, el costo dominante suele venir de comentarios/strings largos y del numero de alternativas por token (`or`).

### Medicion rapida (prueba local)
Se corrio el escaner sobre un archivo Go sintetico grande (`big_example.go`, repeticion de ejemplo):
- Tiempo observado: `real 7.79s` (5000 repeticiones del ejemplo)

Esto nos confirma que el enfoque funciona

## Ejecucion

Comandos de prueba de la version Python secuencial:

```bash
python3 motordelexico.py --spec lexical_spec.sexp --language erlang --input example.erl --output tokens_erlang.txt
python3 motordelexico.py --spec lexical_spec.sexp --language go --input example.go --output tokens_go.txt
python3 motordelexico.py --spec lexical_spec.sexp --language cpp --input example.cpp --output tokens_cpp.txt
```

La version Python tambien acepta multiples archivos de forma secuencial:

```bash
python3 motordelexico.py --spec lexical_spec.sexp --language auto \
  --input example.go --input example.cpp --input example.erl \
  --output tokens_mixtos_python.txt
```

## 7. Nueva version en Go con concurrencia

Implementacion en Go:
- `motordelexico.go`

La aplicacion en Go:
- lee la misma especificacion `lexical_spec.sexp`,
- compila las expresiones-s a regex nativas de Go (`regexp`),
- procesa multiples archivos con goroutines,
- envia cada resultado por un channel,
- evita race conditions dejando que solo la goroutine principal agregue y escriba resultados.

> Nota: como el repositorio contiene archivos C++ de ejemplo en la raiz, se recomienda ejecutar Go por archivo (`go run motordelexico.go`) en vez de `go run .`.

Comandos de prueba:

```bash
go run motordelexico.go --spec lexical_spec.sexp --language go --input example.go --output tokens_go_from_go.txt

go run motordelexico.go --spec lexical_spec.sexp --language auto \
  --input example.go --input example.cpp --input example.erl \
  --output tokens_mixtos_go.txt
```

Tambien puede escribirse un archivo por fuente:

```bash
go run motordelexico.go --spec lexical_spec.sexp --language auto \
  --input example.go --input example.cpp --input example.erl \
  --output-dir out_tokens
```

Pruebas:

```bash
go test motordelexico.go motordelexico_test.go
go test -race motordelexico.go motordelexico_test.go
python3 -c "import ast, pathlib; ast.parse(pathlib.Path('motordelexico.py').read_text(encoding='utf-8'))"
```

### Medicion de tiempos

Para no incluir tiempo de compilacion en Go:

```bash
go build -o /tmp/motordelexico-go motordelexico.go
```

Benchmark usado: dos pasadas sobre `big_example.go`, escribiendo salida combinada.

| Version | Corrida 1 | Corrida 2 | Corrida 3 | Promedio |
| --- | ---: | ---: | ---: | ---: |
| Python secuencial | 15.52s | 15.47s | 15.72s | 15.57s |
| Go concurrente | 0.66s | 0.53s | 0.50s | 0.56s |

La salida generada por ambas versiones fue comparada con `diff` y no presento diferencias.

Archivos de salida generados:
- `tokens_erlang.txt`
- `tokens_go.txt`
- `tokens_cpp.txt`
- opcionalmente, salidas multiarchivo generadas con `--output` o `--output-dir`

Reporte breve:
- `reporte_concurrencia.md`

## Referencias

- Go Specification (lexical elements, keywords, operators, comments): https://go.dev/ref/spec
- Erlang Reference Manual (reserved words): https://www.erlang.org/doc/system/reference_manual.html
- Erlang Reference Manual (comments): https://www.erlang.org/docs/26/reference_manual/modules
- Erlang Reference Manual (atoms y variables): https://www.erlang.org/docs/26/reference_manual/data_types y https://www.erlang.org/docs/26/reference_manual/expressions
- C++ keywords: https://en.cppreference.com/w/cpp/keyword
- C++ comments: https://en.cppreference.com/w/cpp/comments
