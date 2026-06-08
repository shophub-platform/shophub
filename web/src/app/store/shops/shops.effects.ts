import { Injectable, inject } from '@angular/core';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { Router } from '@angular/router';
import { HttpErrorResponse } from '@angular/common/http';
import { catchError, exhaustMap, map, mergeMap, of, tap } from 'rxjs';

import { ShopsActions } from './shops.actions';
import { UiActions } from '../ui/ui.actions';
import { ShopService } from '../../core/services/shop.service';
import { apiError } from '../api-error';

@Injectable()
export class ShopsEffects {
  private readonly actions$ = inject(Actions);
  private readonly shops = inject(ShopService);
  private readonly router = inject(Router);

  load$ = createEffect(() =>
    this.actions$.pipe(
      ofType(ShopsActions.load),
      exhaustMap(() =>
        this.shops.list().pipe(
          map((shops) => ShopsActions.loadSuccess({ shops })),
          catchError((e: HttpErrorResponse) =>
            of(ShopsActions.loadFailure({ error: apiError(e) })),
          ),
        ),
      ),
    ),
  );

  loadOne$ = createEffect(() =>
    this.actions$.pipe(
      ofType(ShopsActions.loadOne),
      exhaustMap(({ id }) =>
        this.shops.get(id).pipe(
          map((shop) => ShopsActions.loadOneSuccess({ shop })),
          catchError((e: HttpErrorResponse) =>
            of(ShopsActions.loadOneFailure({ error: apiError(e) })),
          ),
        ),
      ),
    ),
  );

  create$ = createEffect(() =>
    this.actions$.pipe(
      ofType(ShopsActions.create),
      exhaustMap(({ payload }) =>
        this.shops.create(payload).pipe(
          map((shop) => ShopsActions.createSuccess({ shop })),
          catchError((e: HttpErrorResponse) =>
            of(ShopsActions.createFailure({ error: apiError(e) })),
          ),
        ),
      ),
    ),
  );

  createSuccess$ = createEffect(() =>
    this.actions$.pipe(
      ofType(ShopsActions.createSuccess),
      tap(({ shop }) => this.router.navigate(['/shops', shop.id])),
      map(() =>
        UiActions.showToast({ message: 'Shop created.', kind: 'success' }),
      ),
    ),
  );

  update$ = createEffect(() =>
    this.actions$.pipe(
      ofType(ShopsActions.update),
      exhaustMap(({ id, patch }) =>
        this.shops.update(id, patch).pipe(
          map((shop) => ShopsActions.updateSuccess({ shop })),
          catchError((e: HttpErrorResponse) =>
            of(ShopsActions.updateFailure({ error: apiError(e) })),
          ),
        ),
      ),
    ),
  );

  updateSuccess$ = createEffect(() =>
    this.actions$.pipe(
      ofType(ShopsActions.updateSuccess),
      map(() =>
        UiActions.showToast({ message: 'Configuration saved.', kind: 'success' }),
      ),
    ),
  );

  delete$ = createEffect(() =>
    this.actions$.pipe(
      ofType(ShopsActions.delete),
      exhaustMap(({ id }) =>
        this.shops.remove(id).pipe(
          map(() => ShopsActions.deleteSuccess({ id })),
          catchError((e: HttpErrorResponse) =>
            of(ShopsActions.deleteFailure({ error: apiError(e) })),
          ),
        ),
      ),
    ),
  );

  deleteSuccess$ = createEffect(() =>
    this.actions$.pipe(
      ofType(ShopsActions.deleteSuccess),
      tap(() => this.router.navigate(['/'])),
      map(() =>
        UiActions.showToast({ message: 'Shop deleted.', kind: 'info' }),
      ),
    ),
  );

  failures$ = createEffect(() =>
    this.actions$.pipe(
      ofType(
        ShopsActions.loadFailure,
        ShopsActions.loadOneFailure,
        ShopsActions.createFailure,
        ShopsActions.updateFailure,
        ShopsActions.deleteFailure,
      ),
      mergeMap(({ error }) => of(UiActions.showToast({ message: error, kind: 'error' }))),
    ),
  );
}
