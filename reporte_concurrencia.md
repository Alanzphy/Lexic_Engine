# Reporte breve: versión secuencial y versión concurrente

El programa original en Python se extendió para procesar varios archivos de forma secuencial, reutilizando la misma especificación léxica en `lexical_spec.sexp`. También se desarrolló una nueva versión en Go que lee varios archivos concurrentemente, detecta tokens con regex, envía resultados por channels y evita race conditions al centralizar la escritura de resultados en la goroutine principal.

En cuanto a convenciones, Python mantiene funciones pequeñas, nombres en `snake_case`, uso de `argparse` y separación clara entre parseo de la especificación, escaneo y salida. En Go se usaron estructuras simples, funciones con responsabilidades delimitadas, goroutines para paralelizar archivos independientes y channels para comunicar resultados sin compartir escritura mutable.

La solución conserva el criterio de *longest match*: cuando varios tokens coinciden, se elige el de mayor longitud; si hay empate, se respeta el orden de la especificación. La diferencia principal es que Python evalúa las expresiones regulares con un motor recursivo propio, mientras que Go traduce las expresiones-s a regex del paquete estándar `regexp`, lo que mejora el rendimiento.

## Medición de tiempos

Se realizaron tres ejecuciones usando dos entradas grandes basadas en `big_example.go`.

| Versión | Corrida 1 | Corrida 2 | Corrida 3 | Promedio |
| --- | ---: | ---: | ---: | ---: |
| Python secuencial | 15.52s | 15.47s | 15.72s | 15.57s |
| Go concurrente | 0.66s | 0.53s | 0.50s | 0.56s |

La versión en Go fue más rápida porque combina regex compiladas con procesamiento concurrente de archivos independientes. Además, las salidas de ambas versiones fueron comparadas con `diff` y no presentaron diferencias.

## Conclusión

Separar la especificación léxica del motor facilitó migrar de Python a Go sin redefinir los tokens. Python resultó útil para explicar el algoritmo paso a paso, mientras que Go fue mejor para ejecutar el trabajo en paralelo. El uso de goroutines, channels y agregación centralizada permitió cumplir el objetivo de concurrencia sin introducir condiciones de carrera.
