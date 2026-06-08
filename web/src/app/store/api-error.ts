import { HttpErrorResponse } from '@angular/common/http';

/** Izvlači čitljivu poruku iz backend greške ({"error": "..."}) ili HTTP-a. */
export function apiError(e: HttpErrorResponse): string {
  if (e.error && typeof e.error === 'object' && 'error' in e.error) {
    return String((e.error as { error: unknown }).error);
  }
  if (typeof e.error === 'string' && e.error.trim()) {
    return e.error;
  }
  if (e.status === 0) {
    return 'Server unavailable. Check that the backend is running.';
  }
  return e.message || 'Something went wrong.';
}
