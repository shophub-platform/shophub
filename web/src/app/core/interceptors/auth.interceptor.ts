import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, retry, throwError, timer } from 'rxjs';

import { AuthService } from '../services/auth.service';

const AUTH_FREE = ['/api/v1/auth/login', '/api/v1/auth/register', '/api/v1/auth/refresh'];

/**
 * Funkcionalni interceptor (FZ 4, 9.3):
 *  - dodaje JWT Bearer header (osim na auth endpoint-ima),
 *  - retry logika za GET (mrežni/5xx tranzijenti),
 *  - centralizovano rukovanje greškama; na 401 čisti sesiju i vodi na /login.
 */
export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);
  const router = inject(Router);

  const isAuthFree = AUTH_FREE.some((u) => req.url.includes(u));
  const token = auth.getAccessToken();

  const authReq =
    token && !isAuthFree
      ? req.clone({ setHeaders: { Authorization: `Bearer ${token}` } })
      : req;

  return next(authReq).pipe(
    // Ponovi samo idempotentne GET zahteve, uz blagi backoff.
    retry({
      count: req.method === 'GET' ? 2 : 0,
      delay: (err: HttpErrorResponse, n) => {
        if (err.status && err.status < 500 && err.status !== 0) {
          return throwError(() => err); // 4xx se ne ponavlja
        }
        return timer(300 * n);
      },
    }),
    catchError((err: HttpErrorResponse) => {
      if (err.status === 401 && !isAuthFree) {
        auth.logout();
        router.navigate(['/login']);
      }
      return throwError(() => err);
    }),
  );
};
