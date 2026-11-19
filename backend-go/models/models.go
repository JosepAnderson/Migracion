package models

import "math"

// PeriodoId representa los 4 bimestres escolares
type PeriodoId string

const (
	Bim1 PeriodoId = "bim1"
	Bim2 PeriodoId = "bim2"
	Bim3 PeriodoId = "bim3"
	Bim4 PeriodoId = "bim4"
)

// NotaPeriodo representa las calificaciones de un bimestre y asistencia
type NotaPeriodo struct {
	N1         float64 `json:"n1"`         // Prácticas y tareas (30%)
	N2         float64 `json:"n2"`         // Proyectos y trabajos (30%)
	N3         float64 `json:"n3"`         // Examen bimestral (40%)
	Asistencia float64 `json:"asistencia"` // 0 a 100%
}

// PromedioBimestral calcula la ponderación: N1(30%) + N2(30%) + N3(40%)
func (np NotaPeriodo) PromedioBimestral() float64 {
	prom := (np.N1 * 0.3) + (np.N2 * 0.3) + (np.N3 * 0.4)
	return math.Round(prom*10) / 10
}

// CursoData representa un curso con sus notas por los 4 bimestres
type CursoData struct {
	ID              string                 `json:"id"`
	Nombre          string                 `json:"nombre"`
	Area            string                 `json:"area"`
	Profesor        string                 `json:"profesor"`
	Creditos        int                    `json:"creditos"`
	NotasPorPeriodo map[PeriodoId]NotaPeriodo `json:"notasPorPeriodo"`
}

// PromedioAnual calcula el promedio de los 4 bimestres de este curso
func (c CursoData) PromedioAnual() float64 {
	bims := []PeriodoId{Bim1, Bim2, Bim3, Bim4}
	var suma float64
	for _, b := range bims {
		if nota, ok := c.NotasPorPeriodo[b]; ok {
			suma += nota.PromedioBimestral()
		}
	}
	prom := suma / float64(len(bims))
	return math.Round(prom*10) / 10
}

// AlumnoData representa el estudiante completo con sus datos personales y cursos
type AlumnoData struct {
	ID          string      `json:"id"`
	DNI         string      `json:"dni"`
	Nombres     string      `json:"nombres"`
	Apellidos   string      `json:"apellidos"`
	Grado       string      `json:"grado"`
	Seccion     string      `json:"seccion"`
	AnioLectivo int         `json:"añoLectivo"` // Compatible con el frontend
	Tutor       string      `json:"tutor"`
	Estado      string      `json:"estado"`
	Cursos      []CursoData `json:"cursos"`
}

// PromedioAnualGeneral calcula el promedio de todos los cursos del alumno
func (a AlumnoData) PromedioAnualGeneral() float64 {
	if len(a.Cursos) == 0 {
		return 0
	}
	var suma float64
	for _, c := range a.Cursos {
		suma += c.PromedioAnual()
	}
	prom := suma / float64(len(a.Cursos))
	return math.Round(prom*10) / 10
}

// AsistenciaAnualPromedio calcula el porcentaje general de asistencia anual
func (a AlumnoData) AsistenciaAnualPromedio() float64 {
	if len(a.Cursos) == 0 {
		return 0
	}
	var suma float64
	var count float64
	bims := []PeriodoId{Bim1, Bim2, Bim3, Bim4}
	for _, c := range a.Cursos {
		for _, b := range bims {
			if nota, ok := c.NotasPorPeriodo[b]; ok {
				suma += nota.Asistencia
				count++
			}
		}
	}
	if count == 0 {
		return 0
	}
	return math.Round(suma / count)
}

// RequestActualizarNota payload recibido al calificar un alumno
type RequestActualizarNota struct {
	AlumnoID string      `json:"alumnoId"`
	CursoID  string      `json:"cursoId"`
	Periodo  PeriodoId   `json:"periodo"`
	Notas    NotaPeriodo `json:"notas"`
}

// MetricasResponse resume el estado general para la vista de Inicio y Dashboard
type MetricasResponse struct {
	TotalAlumnos     int     `json:"totalAlumnos"`
	PromedioGeneral  float64 `json:"promedioGeneral"`
	AsistenciaGlobal float64 `json:"asistenciaGlobal"`
	TasaAprobados    float64 `json:"tasaAprobados"`
	AlumnosEnRiesgo  int     `json:"alumnosEnRiesgo"`
}
