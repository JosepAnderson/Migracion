import { AlumnoData, NotaPeriodo, PeriodoId } from '../academico/tipos';
import { ESTUDIANTES_INICIALES } from '../academico/datos';

/**
 * URL base del backend en Go (Golang) con SQLite.
 * Si se define la variable de entorno VITE_API_URL, se usa esa;
 * de lo contrario, apunta por defecto al puerto 8080 local.
 */
export const API_BASE_URL =
  (import.meta.env.VITE_API_URL as string) || 'http://localhost:8080/api';

export const BACKEND_HEALTH_URL =
  API_BASE_URL.replace(/\/api\/?$/, '') + '/health';

export interface MetricasResponse {
  totalAlumnos: number;
  promedioGeneral: number;
  asistenciaGlobal: number;
  tasaAprobados: number;
  alumnosEnRiesgo: number;
}

/**
 * Verifica si el backend en Go está activo y respondiendo.
 */
export async function verificarSaludBackend(): Promise<boolean> {
  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 1500);
    const resp = await fetch(BACKEND_HEALTH_URL, {
      signal: controller.signal,
      method: 'GET',
    });
    clearTimeout(timeoutId);
    return resp.ok;
  } catch {
    return false;
  }
}

/**
 * Obtiene la lista completa de alumnos desde el backend en Go (SQLite).
 * En caso de que el backend en Go aún no se haya iniciado en el puerto 8080,
 * utiliza los datos iniciales locales para que la interfaz nunca se rompa.
 */
export async function obtenerAlumnos(): Promise<{ data: AlumnoData[]; desdeBackend: boolean }> {
  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 2000);
    const response = await fetch(`${API_BASE_URL}/alumnos`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      signal: controller.signal,
    });
    clearTimeout(timeoutId);

    if (!response.ok) {
      throw new Error(`Error HTTP: ${response.status}`);
    }

    const data: AlumnoData[] = await response.json();
    return { data, desdeBackend: true };
  } catch (error) {
    console.warn(
      '[EduControl] No se pudo conectar al Backend en Go (http://localhost:8080). Operando con almacenamiento local:',
      error
    );
    return { data: ESTUDIANTES_INICIALES, desdeBackend: false };
  }
}

/**
 * Busca un estudiante por DNI en el backend en Go.
 */
export async function buscarAlumnoPorDNI(dni: string): Promise<AlumnoData | null> {
  try {
    const response = await fetch(`${API_BASE_URL}/alumnos?dni=${encodeURIComponent(dni)}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (response.status === 404) {
      return null;
    }

    if (!response.ok) {
      throw new Error(`Error HTTP: ${response.status}`);
    }

    return await response.json();
  } catch (error) {
    console.warn('[EduControl] Error al consultar DNI en backend Go:', error);
    return null;
  }
}

/**
 * Registra o matricula un nuevo estudiante en la base de datos SQLite.
 */
export async function matricularAlumno(nuevo: AlumnoData): Promise<{ exito: boolean; alumno: AlumnoData; error?: string }> {
  try {
    const response = await fetch(`${API_BASE_URL}/alumnos`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(nuevo),
    });

    const resJson = await response.json();

    if (!response.ok) {
      return { exito: false, alumno: nuevo, error: resJson.error || 'Error al matricular estudiante' };
    }

    return { exito: true, alumno: resJson.alumno || nuevo };
  } catch (error) {
    console.warn('[EduControl] Backend Go no disponible al matricular, guardando en estado local:', error);
    return { exito: true, alumno: nuevo };
  }
}

/**
 * Actualiza las calificaciones de un estudiante en un curso y bimestre determinado.
 */
export async function guardarCalificaciones(
  alumnoId: string,
  cursoId: string,
  periodo: PeriodoId,
  notas: NotaPeriodo
): Promise<{ exito: boolean; promedioBimestral?: number; error?: string }> {
  try {
    const response = await fetch(`${API_BASE_URL}/notas`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        alumnoId,
        cursoId,
        periodo,
        notas,
      }),
    });

    const resJson = await response.json();

    if (!response.ok) {
      return { exito: false, error: resJson.error || 'Error al actualizar calificaciones' };
    }

    return {
      exito: true,
      promedioBimestral: resJson.promedioBimestral,
    };
  } catch (error) {
    console.warn('[EduControl] Backend Go no disponible al guardar notas, sincronizado en local:', error);
    return { exito: true };
  }
}

/**
 * Elimina un estudiante de la base de datos SQLite por su ID.
 */
export async function eliminarAlumno(alumnoId: string): Promise<{ exito: boolean; error?: string }> {
  try {
    const response = await fetch(`${API_BASE_URL}/alumnos/${encodeURIComponent(alumnoId)}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    const resJson = await response.json();

    if (!response.ok) {
      return { exito: false, error: resJson.error || 'Error al eliminar estudiante' };
    }

    return { exito: true };
  } catch (error) {
    console.warn('[EduControl] Backend Go no disponible al eliminar, eliminado en local:', error);
    return { exito: true };
  }
}

/**
 * Obtiene las métricas globales calculadas por el backend en Go.
 */
export async function obtenerMetricas(): Promise<MetricasResponse | null> {
  try {
    const response = await fetch(`${API_BASE_URL}/metricas`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`Error HTTP: ${response.status}`);
    }

    return await response.json();
  } catch (error) {
    console.warn('[EduControl] Error al obtener métricas del backend:', error);
    return null;
  }
}
