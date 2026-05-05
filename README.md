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

Comandos de prueba:

```bash
python3 motordelexico.py --spec lexical_spec.sexp --language erlang --input example.erl --output tokens_erlang.txt
python3 motordelexico.py --spec lexical_spec.sexp --language go --input example.go --output tokens_go.txt
python3 motordelexico.py --spec lexical_spec.sexp --language cpp --input example.cpp --output tokens_cpp.txt
```

Archivos de salida generados:
- `tokens_erlang.txt`
- `tokens_go.txt`
- `tokens_cpp.txt`

## Referencias

- Go Specification (lexical elements, keywords, operators, comments): https://go.dev/ref/spec
- Erlang Reference Manual (reserved words): https://www.erlang.org/doc/system/reference_manual.html
- Erlang Reference Manual (comments): https://www.erlang.org/docs/26/reference_manual/modules
- Erlang Reference Manual (atoms y variables): https://www.erlang.org/docs/26/reference_manual/data_types y https://www.erlang.org/docs/26/reference_manual/expressions
- C++ keywords: https://en.cppreference.com/w/cpp/keyword
- C++ comments: https://en.cppreference.com/w/cpp/comments
