package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"

	"colegio-backend/database"
	"colegio-backend/models"
)

// Helper para responder en JSON
func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Helper para responder con error JSON
func jsonError(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}

// HandleAlumnos maneja /api/alumnos (GET todos o POST nuevo)
func HandleAlumnos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Verificar si hay filtro por DNI: /api/alumnos?dni=72458912
		dni := r.URL.Query().Get("dni")
		if strings.TrimSpace(dni) != "" {
			alumno, err := database.ObtenerAlumnoPorDNI(strings.TrimSpace(dni))
			if err != nil {
				jsonError(w, http.StatusNotFound, fmt.Sprintf("No se encontró ningún estudiante con el DNI %s", dni))
				return
			}
			jsonResponse(w, http.StatusOK, alumno)
			return
		}

		// Si no hay DNI, retornar lista completa
		alumnos, err := database.ObtenerTodosLosAlumnos()
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Error al consultar los alumnos: "+err.Error())
			return
		}
		if alumnos == nil {
			alumnos = []models.AlumnoData{}
		}
		jsonResponse(w, http.StatusOK, alumnos)

	case http.MethodPost:
		var nuevo models.AlumnoData
		if err := json.NewDecoder(r.Body).Decode(&nuevo); err != nil {
			jsonError(w, http.StatusBadRequest, "Datos del estudiante inválidos")
			return
		}

		if strings.TrimSpace(nuevo.DNI) == "" || strings.TrimSpace(nuevo.Nombres) == "" || strings.TrimSpace(nuevo.Apellidos) == "" {
			jsonError(w, http.StatusBadRequest, "DNI, nombres y apellidos son campos obligatorios")
			return
		}

		// Validar si ya existe el DNI
		existente, _ := database.ObtenerAlumnoPorDNI(nuevo.DNI)
		if existente != nil {
			jsonError(w, http.StatusConflict, fmt.Sprintf("Ya existe un alumno registrado con el DNI %s", nuevo.DNI))
			return
		}

		// Si no tiene ID, generarlo
		if nuevo.ID == "" {
			nuevo.ID = fmt.Sprintf("est-%d", nuevo.AnioLectivo%1000)
		}
		if nuevo.AnioLectivo == 0 {
			nuevo.AnioLectivo = 2026
		}
		if nuevo.Estado == "" {
			nuevo.Estado = "Activo"
		}

		// Si no tiene cursos configurados, cargar los 3 cursos por defecto
		if len(nuevo.Cursos) == 0 {
			cursosBase, err := database.ObtenerTodosLosCursos()
			if err == nil {
				bims := []models.PeriodoId{models.Bim1, models.Bim2, models.Bim3, models.Bim4}
				for i := range cursosBase {
					cursosBase[i].NotasPorPeriodo = make(map[models.PeriodoId]models.NotaPeriodo)
					for _, b := range bims {
						cursosBase[i].NotasPorPeriodo[b] = models.NotaPeriodo{N1: 0, N2: 0, N3: 0, Asistencia: 100}
					}
				}
				nuevo.Cursos = cursosBase
			}
		}

		if err := database.CrearAlumno(nuevo); err != nil {
			jsonError(w, http.StatusInternalServerError, "Error al registrar el alumno: "+err.Error())
			return
		}

		jsonResponse(w, http.StatusCreated, map[string]interface{}{
			"mensaje": "Estudiante matriculado con éxito",
			"alumno":  nuevo,
		})

	default:
		jsonError(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

// HandleAlumnoPorID maneja /api/alumnos/{id} (GET por ID, DELETE)
func HandleAlumnoPorID(w http.ResponseWriter, r *http.Request) {
	// Extraer ID de la ruta: /api/alumnos/est-001
	path := strings.TrimPrefix(r.URL.Path, "/api/alumnos/")
	partes := strings.Split(path, "/")
	alumnoID := partes[0]

	if alumnoID == "" {
		jsonError(w, http.StatusBadRequest, "ID de estudiante requerido")
		return
	}

	switch r.Method {
	case http.MethodGet:
		alumno, err := database.ObtenerAlumnoPorID(alumnoID)
		if err != nil {
			jsonError(w, http.StatusNotFound, "Estudiante no encontrado")
			return
		}
		jsonResponse(w, http.StatusOK, alumno)

	case http.MethodDelete:
		err := database.EliminarAlumno(alumnoID)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Error al eliminar el estudiante: "+err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, map[string]string{
			"mensaje": fmt.Sprintf("Estudiante con ID %s eliminado correctamente", alumnoID),
		})

	default:
		jsonError(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

// HandleActualizarNotas maneja PUT /api/notas para calificar a un estudiante
func HandleActualizarNotas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "Método debe ser PUT o POST")
		return
	}

	var req models.RequestActualizarNota
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "Formato JSON de calificación inválido")
		return
	}

	if req.AlumnoID == "" || req.CursoID == "" || req.Periodo == "" {
		jsonError(w, http.StatusBadRequest, "alumnoId, cursoId y periodo son requeridos")
		return
	}

	// Validar rango de notas (0 a 20)
	if req.Notas.N1 < 0 || req.Notas.N1 > 20 ||
		req.Notas.N2 < 0 || req.Notas.N2 > 20 ||
		req.Notas.N3 < 0 || req.Notas.N3 > 20 {
		jsonError(w, http.StatusBadRequest, "Las notas deben estar en el rango de 0 a 20")
		return
	}

	if req.Notas.Asistencia < 0 || req.Notas.Asistencia > 100 {
		req.Notas.Asistencia = 100
	}

	if err := database.GuardarNota(req.AlumnoID, req.CursoID, req.Periodo, req.Notas); err != nil {
		jsonError(w, http.StatusInternalServerError, "Error al guardar la calificación: "+err.Error())
		return
	}

	promedioCalculado := req.Notas.PromedioBimestral()

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"mensaje":           "Calificaciones actualizadas exitosamente",
		"promedioBimestral": promedioCalculado,
		"aprobado":          promedioCalculado >= 10.5,
		"periodo":           req.Periodo,
	})
}

// HandleMetricas maneja GET /api/metricas calculando promedios globales
func HandleMetricas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	alumnos, err := database.ObtenerTodosLosAlumnos()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Error al calcular métricas: "+err.Error())
		return
	}

	total := len(alumnos)
	if total == 0 {
		jsonResponse(w, http.StatusOK, models.MetricasResponse{
			TotalAlumnos:     0,
			PromedioGeneral:  0,
			AsistenciaGlobal: 0,
			TasaAprobados:    0,
			AlumnosEnRiesgo:  0,
		})
		return
	}

	var sumaPromedios float64
	var sumaAsistencias float64
	aprobados := 0
	enRiesgo := 0

	for _, a := range alumnos {
		prom := a.PromedioAnualGeneral()
		asist := a.AsistenciaAnualPromedio()
		sumaPromedios += prom
		sumaAsistencias += asist

		if prom >= 11 {
			aprobados++
		} else {
			enRiesgo++
		}
	}

	promGeneral := math.Round((sumaPromedios/float64(total))*10) / 10
	asistGlobal := math.Round(sumaAsistencias / float64(total))
	tasaAprobados := math.Round((float64(aprobados) / float64(total)) * 100)

	jsonResponse(w, http.StatusOK, models.MetricasResponse{
		TotalAlumnos:     total,
		PromedioGeneral:  promGeneral,
		AsistenciaGlobal: asistGlobal,
		TasaAprobados:    tasaAprobados,
		AlumnosEnRiesgo:  enRiesgo,
	})
}

// MiddlewareCORS permite la conexión desde React (puerto 3000 o 5173) y Angular (puerto 4200)
func MiddlewareCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
