# Backend en Go (Golang) + SQLite — Sistema de Gestión Académica

Este backend provee una **API REST rápida, concurrente y ligera** desarrollada en Go (Golang) para el sistema de control de notas y consulta por DNI. Utiliza **SQLite** como base de datos embebida, por lo que **no necesitas instalar MySQL, PostgreSQL ni configurar contraseñas**.

---

## 📁 Estructura del Proyecto

```text
backend-go/
├── main.go                # Punto de entrada del servidor HTTP (puerto :8080)
├── go.mod                 # Definición del módulo y dependencias en Go
├── database/
│   └── db.go              # Conexión a SQLite (colegio.db), creación de tablas y datos iniciales
├── models/
│   └── models.go          # Estructuras de datos (Alumno, Notas, Cursos) y fórmulas de promedio
├── handlers/
│   └── handlers.go        # Controladores REST (GET, POST, PUT, DELETE) y Middleware CORS
└── README.md              # Documentación y guía de sustentación
```

---

## 🚀 ¿Cómo Ejecutarlo en tu Computadora (Windows / Linux / Mac)?

1. Abre tu terminal o PowerShell.
2. Ingresa a la carpeta del backend:
   ```powershell
   cd backend-go
   ```
3. Descarga las dependencias automáticamente:
   ```powershell
   go mod tidy
   ```
4. Inicia el servidor:
   ```powershell
   go run main.go
   ```

El servidor creará automáticamente el archivo `colegio.db` en la carpeta y precargará los 5 alumnos con sus calificaciones de prueba. Verás este mensaje en la consola:

```text
================================================================
   SISTEMA DE GESTIÓN ACADÉMICA - BACKEND EN GO (GOLANG)        
================================================================
✓ Servidor HTTP activo en:       http://localhost:8080
✓ Archivo de base de datos:     colegio.db (SQLite)
----------------------------------------------------------------
Endpoints disponibles:
  • GET    http://localhost:8080/api/alumnos
  • GET    http://localhost:8080/api/alumnos?dni=72458912
  • POST   http://localhost:8080/api/alumnos
  • PUT    http://localhost:8080/api/notas
  • DELETE http://localhost:8080/api/alumnos/{id}
  • GET    http://localhost:8080/api/metricas
================================================================
```

---

## 📡 Endpoints de la API REST

| Método | Endpoint | Descripción |
|---|---|---|
| `GET` | `/api/alumnos` | Lista todos los alumnos registrados con sus notas y cursos. |
| `GET` | `/api/alumnos?dni=72458912` | Busca a un alumno por su número de DNI para la boleta de notas. |
| `GET` | `/api/alumnos/{id}` | Obtiene los datos detallados de un alumno por su ID. |
| `POST` | `/api/alumnos` | Registra (matricula) a un nuevo alumno en la base de datos. |
| `PUT` | `/api/notas` | Actualiza las notas (N1, N2, N3 y Asistencia) de un bimestre. |
| `DELETE` | `/api/alumnos/{id}` | Elimina al estudiante y todas sus notas en cascada. |
| `GET` | `/api/metricas` | Devuelve el total de alumnos, promedio general y % de aprobados. |
| `GET` | `/health` | Verifica que el servidor esté activo y respondiendo. |

---

## 🗄️ Esquema de la Base de Datos SQLite

1. **Tabla `cursos`**:
   - `id` (TEXT, PK): `'mat'`, `'cyt'`, `'com'`
   - `nombre`, `area`, `profesor`, `creditos`

2. **Tabla `alumnos`**:
   - `id` (TEXT, PK), `dni` (TEXT, UNIQUE), `nombres`, `apellidos`, `grado`, `seccion`, `anio_lectivo`, `tutor`, `estado`

3. **Tabla `notas`**:
   - `id` (INTEGER, PK AUTOINCREMENT)
   - `alumno_id` (FK alumnos), `curso_id` (FK cursos), `periodo_id` (`'bim1'`, `'bim2'`, `'bim3'`, `'bim4'`)
   - `n1` (30%), `n2` (30%), `n3` (40%), `asistencia` (0-100%)
   - Restricción única: `UNIQUE(alumno_id, curso_id, periodo_id)`

---

## 🎓 Preguntas Clave para Exponer a tu Profesor

1. **¿Por qué usaste Go para el backend?**
   * *Respuesta:* Porque Go compila directamente a código máquina nativo, tiene un consumo mínimo de memoria RAM y gestiona múltiples peticiones concurrentes de manera ultra eficiente gracias a sus goroutines.

2. **¿Por qué SQLite y no MySQL?**
   * *Respuesta:* SQLite es una base de datos relacional serverless y autocontenida que se almacena en un solo archivo (`colegio.db`). Es ideal para despliegues portátiles y rápidos sin requerir servidores de base de datos externos pesados.

3. **¿Cómo se comunica el Frontend con el Backend?**
   * *Respuesta:* A través del protocolo HTTP mediante peticiones JSON (REST). El backend incluye un Middleware CORS (`MiddlewareCORS`) que autoriza al frontend (en React o Angular) a enviar peticiones seguras.

4. **¿Dónde se calcula el promedio ponderado?**
   * *Respuesta:* En `models/models.go` con la función `PromedioBimestral()` aplicando la fórmula institucional: `(N1 * 0.3) + (N2 * 0.3) + (N3 * 0.4)`.
