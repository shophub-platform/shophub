import { Injectable, inject } from '@angular/core';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { MatSnackBar } from '@angular/material/snack-bar';
import { tap } from 'rxjs';

import { UiActions } from './ui.actions';

/** Toast notifikacije preko MatSnackBar (FZ 4, 9.1). */
@Injectable()
export class UiEffects {
  private readonly actions$ = inject(Actions);
  private readonly snack = inject(MatSnackBar);

  showToast$ = createEffect(
    () =>
      this.actions$.pipe(
        ofType(UiActions.showToast),
        tap(({ message, kind }) => {
          this.snack.open(message, '✕', {
            duration: 4000,
            horizontalPosition: 'center',
            verticalPosition: 'bottom',
            panelClass: [`toast-${kind ?? 'info'}`],
          });
        }),
      ),
    { dispatch: false },
  );
}
