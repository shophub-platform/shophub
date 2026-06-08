import { Injectable, inject } from '@angular/core';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { Router } from '@angular/router';
import { HttpErrorResponse } from '@angular/common/http';
import { catchError, exhaustMap, map, of, tap } from 'rxjs';

import { AuthActions } from './auth.actions';
import { UiActions } from '../ui/ui.actions';
import { AuthService } from '../../core/services/auth.service';
import { apiError } from '../api-error';

@Injectable()
export class AuthEffects {
  private readonly actions$ = inject(Actions);
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  login$ = createEffect(() =>
    this.actions$.pipe(
      ofType(AuthActions.login),
      exhaustMap(({ creds, redirect }) =>
        this.auth.login(creds).pipe(
          map((tokens) => AuthActions.loginSuccess({ tokens, redirect })),
          catchError((e: HttpErrorResponse) =>
            of(AuthActions.loginFailure({ error: apiError(e) })),
          ),
        ),
      ),
    ),
  );

  register$ = createEffect(() =>
    this.actions$.pipe(
      ofType(AuthActions.register),
      exhaustMap(({ payload }) =>
        this.auth.register(payload).pipe(
          map((tokens) => AuthActions.registerSuccess({ tokens })),
          catchError((e: HttpErrorResponse) =>
            of(AuthActions.registerFailure({ error: apiError(e) })),
          ),
        ),
      ),
    ),
  );

  // Posle uspešne prijave/registracije: navigacija + toast.
  authSuccessNav$ = createEffect(
    () =>
      this.actions$.pipe(
        ofType(AuthActions.loginSuccess, AuthActions.registerSuccess),
        tap((a) => {
          const redirect = 'redirect' in a ? a.redirect : undefined;
          this.router.navigateByUrl(redirect || '/');
        }),
      ),
    { dispatch: false },
  );

  authSuccessToast$ = createEffect(() =>
    this.actions$.pipe(
      ofType(AuthActions.loginSuccess, AuthActions.registerSuccess),
      map(() => UiActions.showToast({ message: 'Welcome!', kind: 'success' })),
    ),
  );

  authFailureToast$ = createEffect(() =>
    this.actions$.pipe(
      ofType(AuthActions.loginFailure, AuthActions.registerFailure),
      map(({ error }) => UiActions.showToast({ message: error, kind: 'error' })),
    ),
  );

  logout$ = createEffect(
    () =>
      this.actions$.pipe(
        ofType(AuthActions.logout),
        tap(() => {
          this.auth.logout();
          this.router.navigate(['/login']);
        }),
      ),
    { dispatch: false },
  );
}
