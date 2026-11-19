package database

import (
	"database/sql"
	"fmt"
	"log"

	"colegio-backend/models"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// InitDB inicializa la base de datos SQLite y crea las tablas si no existen
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error al abrir la base de datos SQLite: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error al conectar con SQLite: %w", err)
	}

	DB = db

	// Crear las tablas necesarias
	if err := crearTablas(); err != nil {
		return nil, fmt.Errorf("error al crear tablas: %w", err)
	}

	// Insertar datos iniciales si la base de datos está vacía
	if err := poblarDatosIniciales(); err != nil {
		log.Printf("Advertencia al poblar datos iniciales: %v\n", err)
	}

	return DB, nil
}

func crearTablas() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS cursos (
			id TEXT PRIMARY KEY,
			nombre TEXT NOT NULL,
			area TEXT NOT NULL,
			profesor TEXT NOT NULL,
			creditos INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS alumnos (
			id TEXT PRIMARY KEY,
			dni TEXT UNIQUE NOT NULL,
			nombres TEXT NOT NULL,
			apellidos TEXT NOT NULL,
			grado TEXT NOT NULL,
			seccion TEXT NOT NULL,
			anio_lectivo INTEGER NOT NULL,
			tutor TEXT NOT NULL,
			estado TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS notas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			alumno_id TEXT NOT NULL,
			curso_id TEXT NOT NULL,
			periodo_id TEXT NOT NULL,
			n1 REAL NOT NULL,
			n2 REAL NOT NULL,
			n3 REAL NOT NULL,
			asistencia REAL NOT NULL,
			UNIQUE(alumno_id, curso_id, periodo_id),
			FOREIGN KEY (alumno_id) REFERENCES alumnos(id) ON DELETE CASCADE,
			FOREIGN KEY (curso_id) REFERENCES cursos(id) ON DELETE CASCADE
		);`,
	}

	for _, query := range queries {
		if _, err := DB.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

// ObtenerTodosLosCursos retorna los 3 cursos base
func ObtenerTodosLosCursos() ([]models.CursoData, error) {
	rows, err := DB.Query("SELECT id, nombre, area, profesor, creditos FROM cursos ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cursos []models.CursoData
	for rows.Next() {
		var c models.CursoData
		if err := rows.Scan(&c.ID, &c.Nombre, &c.Area, &c.Profesor, &c.Creditos); err != nil {
			return nil, err
		}
		c.NotasPorPeriodo = make(map[models.PeriodoId]models.NotaPeriodo)
		cursos = append(cursos, c)
	}
	return cursos, nil
}

// ObtenerTodosLosAlumnos obtiene la lista completa de alumnos con sus notas
func ObtenerTodosLosAlumnos() ([]models.AlumnoData, error) {
	rows, err := DB.Query("SELECT id, dni, nombres, apellidos, grado, seccion, anio_lectivo, tutor, estado FROM alumnos ORDER BY apellidos ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alumnos []models.AlumnoData
	for rows.Next() {
		var a models.AlumnoData
		if err := rows.Scan(&a.ID, &a.DNI, &a.Nombres, &a.Apellidos, &a.Grado, &a.Seccion, &a.AnioLectivo, &a.Tutor, &a.Estado); err != nil {
			return nil, err
		}

		cursos, err := CargarCursosYNotasPorAlumno(a.ID)
		if err != nil {
			return nil, err
		}
		a.Cursos = cursos
		alumnos = append(alumnos, a)
	}

	return alumnos, nil
}

// ObtenerAlumnoPorDNI busca un alumno por su DNI único
func ObtenerAlumnoPorDNI(dni string) (*models.AlumnoData, error) {
	var a models.AlumnoData
	row := DB.QueryRow("SELECT id, dni, nombres, apellidos, grado, seccion, anio_lectivo, tutor, estado FROM alumnos WHERE dni = ?", dni)
	err := row.Scan(&a.ID, &a.DNI, &a.Nombres, &a.Apellidos, &a.Grado, &a.Seccion, &a.AnioLectivo, &a.Tutor, &a.Estado)
	if err != nil {
		return nil, err
	}

	cursos, err := CargarCursosYNotasPorAlumno(a.ID)
	if err != nil {
		return nil, err
	}
	a.Cursos = cursos
	return &a, nil
}

// ObtenerAlumnoPorID busca un alumno por su ID
func ObtenerAlumnoPorID(id string) (*models.AlumnoData, error) {
	var a models.AlumnoData
	row := DB.QueryRow("SELECT id, dni, nombres, apellidos, grado, seccion, anio_lectivo, tutor, estado FROM alumnos WHERE id = ?", id)
	err := row.Scan(&a.ID, &a.DNI, &a.Nombres, &a.Apellidos, &a.Grado, &a.Seccion, &a.AnioLectivo, &a.Tutor, &a.Estado)
	if err != nil {
		return nil, err
	}

	cursos, err := CargarCursosYNotasPorAlumno(a.ID)
	if err != nil {
		return nil, err
	}
	a.Cursos = cursos
	return &a, nil
}

// CargarCursosYNotasPorAlumno arma la estructura de cursos y notas para un alumno
func CargarCursosYNotasPorAlumno(alumnoID string) ([]models.CursoData, error) {
	cursosBase, err := ObtenerTodosLosCursos()
	if err != nil {
		return nil, err
	}

	// Consultar todas las notas del alumno
	rows, err := DB.Query("SELECT curso_id, periodo_id, n1, n2, n3, asistencia FROM notas WHERE alumno_id = ?", alumnoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notasMap := make(map[string]map[models.PeriodoId]models.NotaPeriodo)
	for rows.Next() {
		var cursoID string
		var periodo models.PeriodoId
		var n models.NotaPeriodo
		if err := rows.Scan(&cursoID, &periodo, &n.N1, &n.N2, &n.N3, &n.Asistencia); err != nil {
			return nil, err
		}
		if _, ok := notasMap[cursoID]; !ok {
			notasMap[cursoID] = make(map[models.PeriodoId]models.NotaPeriodo)
		}
		notasMap[cursoID][periodo] = n
	}

	// Asignar las notas a cada curso (con valores por defecto si no existen)
	bims := []models.PeriodoId{models.Bim1, models.Bim2, models.Bim3, models.Bim4}
	for i := range cursosBase {
		cid := cursosBase[i].ID
		cursosBase[i].NotasPorPeriodo = make(map[models.PeriodoId]models.NotaPeriodo)
		for _, b := range bims {
			if nota, existe := notasMap[cid][b]; existe {
				cursosBase[i].NotasPorPeriodo[b] = nota
			} else {
				cursosBase[i].NotasPorPeriodo[b] = models.NotaPeriodo{N1: 0, N2: 0, N3: 0, Asistencia: 100}
			}
		}
	}

	return cursosBase, nil
}

// GuardarNota actualiza o inserta las notas de un periodo
func GuardarNota(alumnoID, cursoID string, periodo models.PeriodoId, n models.NotaPeriodo) error {
	query := `
		INSERT INTO notas (alumno_id, curso_id, periodo_id, n1, n2, n3, asistencia)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(alumno_id, curso_id, periodo_id) DO UPDATE SET
			n1 = excluded.n1,
			n2 = excluded.n2,
			n3 = excluded.n3,
			asistencia = excluded.asistencia;
	`
	_, err := DB.Exec(query, alumnoID, cursoID, string(periodo), n.N1, n.N2, n.N3, n.Asistencia)
	return err
}

// CrearAlumno matricula un nuevo alumno en la base de datos
func CrearAlumno(a models.AlumnoData) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"INSERT INTO alumnos (id, dni, nombres, apellidos, grado, seccion, anio_lectivo, tutor, estado) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		a.ID, a.DNI, a.Nombres, a.Apellidos, a.Grado, a.Seccion, a.AnioLectivo, a.Tutor, a.Estado,
	)
	if err != nil {
		return err
	}

	// Inicializar notas en 0 para los 3 cursos y 4 bimestres
	bims := []models.PeriodoId{models.Bim1, models.Bim2, models.Bim3, models.Bim4}
	for _, curso := range a.Cursos {
		for _, b := range bims {
			nota := curso.NotasPorPeriodo[b]
			if nota.Asistencia == 0 {
				nota.Asistencia = 100
			}
			_, err = tx.Exec(
				"INSERT INTO notas (alumno_id, curso_id, periodo_id, n1, n2, n3, asistencia) VALUES (?, ?, ?, ?, ?, ?, ?)",
				a.ID, curso.ID, string(b), nota.N1, nota.N2, nota.N3, nota.Asistencia,
			)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// EliminarAlumno borra un alumno y todas sus calificaciones asociadas
func EliminarAlumno(id string) error {
	_, err := DB.Exec("DELETE FROM alumnos WHERE id = ?", id)
	return err
}

// poblarDatosIniciales llena la base de datos con los cursos y 5 alumnos de muestra
func poblarDatosIniciales() error {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM cursos").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		cursosIniciales := []models.CursoData{
			{ID: "mat", Nombre: "Matemáticas", Area: "Ciencias Exactas", Profesor: "Mg. Carlos Paredes", Creditos: 5},
			{ID: "cyt", Nombre: "Ciencias", Area: "Ciencias Naturales", Profesor: "Dr. Roberto Huamán", Creditos: 4},
			{ID: "com", Nombre: "Comunicación", Area: "Humanidades", Profesor: "Lic. Martha Gómez", Creditos: 5},
		}
		for _, c := range cursosIniciales {
			_, err := DB.Exec("INSERT INTO cursos (id, nombre, area, profesor, creditos) VALUES (?, ?, ?, ?, ?)", c.ID, c.Nombre, c.Area, c.Profesor, c.Creditos)
			if err != nil {
				return err
			}
		}
	}

	var countAlumnos int
	err = DB.QueryRow("SELECT COUNT(*) FROM alumnos").Scan(&countAlumnos)
	if err != nil {
		return err
	}

	if countAlumnos == 0 {
		alumnosIniciales := []models.AlumnoData{
			{
				ID: "est-001", DNI: "72458912", Nombres: "Carlos Daniel", Apellidos: "Mendoza Rojas",
				Grado: "4° de Secundaria", Seccion: "A", AnioLectivo: 2026, Tutor: "Lic. Martha Gómez", Estado: "Activo",
			},
			{
				ID: "est-002", DNI: "74125896", Nombres: "Valeria Sofía", Apellidos: "Quispe Vargas",
				Grado: "4° de Secundaria", Seccion: "A", AnioLectivo: 2026, Tutor: "Lic. Martha Gómez", Estado: "Activo",
			},
			{
				ID: "est-003", DNI: "75315984", Nombres: "Diego Alejandro", Apellidos: "Flores Castro",
				Grado: "4° de Secundaria", Seccion: "A", AnioLectivo: 2026, Tutor: "Lic. Martha Gómez", Estado: "Activo",
			},
			{
				ID: "est-004", DNI: "76985214", Nombres: "Luciana Belén", Apellidos: "Herrera Morales",
				Grado: "4° de Secundaria", Seccion: "A", AnioLectivo: 2026, Tutor: "Lic. Martha Gómez", Estado: "Activo",
			},
			{
				ID: "est-005", DNI: "71234567", Nombres: "Mateo Sebastián", Apellidos: "Castillo Ruiz",
				Grado: "4° de Secundaria", Seccion: "A", AnioLectivo: 2026, Tutor: "Lic. Martha Gómez", Estado: "Activo",
			},
		}

		for _, a := range alumnosIniciales {
			_, err := DB.Exec("INSERT INTO alumnos (id, dni, nombres, apellidos, grado, seccion, anio_lectivo, tutor, estado) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
				a.ID, a.DNI, a.Nombres, a.Apellidos, a.Grado, a.Seccion, a.AnioLectivo, a.Tutor, a.Estado)
			if err != nil {
				return err
			}
		}

		// Insertar notas de muestra
		notasMuestra := []struct {
			aid, cid, pid string
			n1, n2, n3, as float64
		}{
			// est-001
			{"est-001", "mat", "bim1", 17, 16, 18, 96},
			{"est-001", "mat", "bim2", 18, 17, 19, 98},
			{"est-001", "mat", "bim3", 16, 18, 17, 95},
			{"est-001", "mat", "bim4", 19, 18, 19, 99},
			{"est-001", "cyt", "bim1", 15, 16, 17, 92},
			{"est-001", "cyt", "bim2", 16, 16, 18, 94},
			{"est-001", "cyt", "bim3", 17, 18, 18, 96},
			{"est-001", "cyt", "bim4", 18, 19, 19, 98},
			{"est-001", "com", "bim1", 16, 17, 16, 95},
			{"est-001", "com", "bim2", 17, 18, 17, 97},
			{"est-001", "com", "bim3", 18, 17, 18, 98},
			{"est-001", "com", "bim4", 19, 19, 20, 100},

			// est-002
			{"est-002", "mat", "bim1", 19, 18, 20, 99},
			{"est-002", "mat", "bim2", 18, 19, 19, 98},
			{"est-002", "mat", "bim3", 20, 19, 20, 100},
			{"est-002", "mat", "bim4", 19, 20, 20, 100},
			{"est-002", "cyt", "bim1", 17, 18, 18, 96},
			{"est-002", "cyt", "bim2", 18, 19, 18, 97},
			{"est-002", "cyt", "bim3", 19, 19, 20, 99},
			{"est-002", "cyt", "bim4", 20, 19, 20, 100},
			{"est-002", "com", "bim1", 18, 19, 19, 98},
			{"est-002", "com", "bim2", 19, 19, 20, 99},
			{"est-002", "com", "bim3", 19, 20, 20, 100},
			{"est-002", "com", "bim4", 20, 20, 20, 100},

			// est-003
			{"est-003", "mat", "bim1", 11, 10, 12, 85},
			{"est-003", "mat", "bim2", 12, 11, 13, 86},
			{"est-003", "mat", "bim3", 10, 12, 11, 84},
			{"est-003", "mat", "bim4", 13, 12, 14, 88},
			{"est-003", "cyt", "bim1", 13, 14, 12, 88},
			{"est-003", "cyt", "bim2", 14, 13, 14, 89},
			{"est-003", "cyt", "bim3", 12, 13, 13, 87},
			{"est-003", "cyt", "bim4", 14, 14, 15, 90},
			{"est-003", "com", "bim1", 12, 13, 11, 86},
			{"est-003", "com", "bim2", 13, 12, 13, 87},
			{"est-003", "com", "bim3", 11, 12, 12, 85},
			{"est-003", "com", "bim4", 13, 13, 14, 89},

			// est-004
			{"est-004", "mat", "bim1", 15, 14, 16, 91},
			{"est-004", "mat", "bim2", 16, 15, 16, 93},
			{"est-004", "mat", "bim3", 15, 16, 17, 92},
			{"est-004", "mat", "bim4", 17, 16, 18, 95},
			{"est-004", "cyt", "bim1", 16, 17, 16, 94},
			{"est-004", "cyt", "bim2", 17, 16, 17, 95},
			{"est-004", "cyt", "bim3", 18, 17, 18, 96},
			{"est-004", "cyt", "bim4", 18, 18, 19, 97},
			{"est-004", "com", "bim1", 14, 15, 15, 90},
			{"est-004", "com", "bim2", 15, 15, 16, 92},
			{"est-004", "com", "bim3", 16, 16, 17, 93},
			{"est-004", "com", "bim4", 17, 17, 17, 94},

			// est-005
			{"est-005", "mat", "bim1", 13, 14, 15, 88},
			{"est-005", "mat", "bim2", 14, 14, 15, 90},
			{"est-005", "mat", "bim3", 15, 15, 16, 92},
			{"est-005", "mat", "bim4", 16, 16, 17, 94},
			{"est-005", "cyt", "bim1", 12, 13, 14, 87},
			{"est-005", "cyt", "bim2", 13, 14, 14, 89},
			{"est-005", "cyt", "bim3", 14, 15, 15, 91},
			{"est-005", "cyt", "bim4", 15, 15, 16, 93},
			{"est-005", "com", "bim1", 14, 14, 15, 90},
			{"est-005", "com", "bim2", 15, 15, 16, 92},
			{"est-005", "com", "bim3", 15, 16, 16, 93},
			{"est-005", "com", "bim4", 16, 17, 17, 95},
		}

		for _, nm := range notasMuestra {
			_, err := DB.Exec(
				"INSERT INTO notas (alumno_id, curso_id, periodo_id, n1, n2, n3, asistencia) VALUES (?, ?, ?, ?, ?, ?, ?)",
				nm.aid, nm.cid, nm.pid, nm.n1, nm.n2, nm.n3, nm.as,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
