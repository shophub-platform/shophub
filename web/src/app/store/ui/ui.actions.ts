import { createActionGroup, emptyProps, props } from '@ngrx/store';

export type ToastKind = 'success' | 'error' | 'info';

export const UiActions = createActionGroup({
  source: 'UI',
  events: {
    'Show Toast': props<{ message: string; kind?: ToastKind }>(),
    'Set Loading': props<{ loading: boolean }>(),
    'Clear Error': emptyProps(),
  },
});
