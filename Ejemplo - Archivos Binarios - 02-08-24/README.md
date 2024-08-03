# Cambios realizados 

## Uso de `int64` en lugar de `int`

Cambo de:  tipo de los campos de `int` a `int64`. Esto se hace para asegurar la compatibilidad y consistencia al escribir y leer datos binarios, ya que `int` puede tener tamaños diferentes dependiendo de la plataforma (32 bits o 64 bits), mientras que `int64` tiene un tamaño fijo de 64 bits en todas las plataformas. Esto es importante para garantizar que los datos se escriban y lean correctamente sin importar el entorno de ejecución.

## Agregado de `Tipo` en las estructuras

Se ha agregado un campo `Tipo` tanto en la estructura `Profesor` como en la estructura `Estudiante`. Este campo se utiliza para diferenciar entre los dos tipos de registros en el archivo binario. El valor `1` indica un `Profesor` y el valor `2` indica un `Estudiante`. 

## Escritura y lectura de datos

En el código antiguo dado en lab, los datos se escribían y leían utilizando la biblioteca `encoding/gob`. En el código nuevo, se utiliza la biblioteca `encoding/binary` para una gestión más precisa y controlada de los datos binarios. 

- **Escritura de datos:**
    - En el código nuevo, primero se escribe el campo `Tipo`, seguido del `ID` y luego la longitud de las cadenas `Nombre` y `Apellido`, seguidas de las propias cadenas.
    - Esto garantiza que al leer los datos, se puede saber exactamente cuántos bytes leer para cada campo, evitando problemas con tamaños variables de cadenas.

- **Lectura de datos:**
    - En el código nuevo, se lee el campo `Tipo` primero para determinar si el siguiente registro es un `Profesor` o un `Estudiante`.
    - Luego, se leen el `ID`, la longitud de la cadena `Nombre`, la cadena `Nombre`, la longitud de la cadena `Apellido`, y la cadena `Apellido` en ese orden.

## Diferencias en el código

### Código viejo

- Uso de `int` para campos numéricos.
- No se distingue entre `Profesor` y `Estudiante` usando un campo `Tipo`.
- Uso de `encoding/gob` para escribir y leer datos binarios, lo que puede ser menos preciso en términos de control de tamaño de los datos y compatibilidad entre diferentes plataformas.
- Escritura y lectura de datos sin un tamaño fijo para cadenas, lo que puede llevar a inconsistencias.

### Código nuevo

- Cambio de `int` a `int64` para asegurar consistencia en el tamaño de los datos en todas las plataformas.
- Adición de un campo `Tipo` para diferenciar entre `Profesor` y `Estudiante`.
- Uso de `encoding/binary` para escribir y leer datos binarios, proporcionando un mayor control sobre el formato de los datos.
- Escritura y lectura de la longitud de las cadenas antes de las cadenas mismas, asegurando que se sepa exactamente cuántos bytes leer para cada campo.
