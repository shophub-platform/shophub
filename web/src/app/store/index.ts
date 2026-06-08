import { ActionReducerMap } from '@ngrx/store';

import { authReducer, AuthState } from './auth/auth.reducer';
import { shopsReducer, ShopsState } from './shops/shops.reducer';
import { uiReducer, UiState } from './ui/ui.reducer';

export interface AppState {
  auth: AuthState;
  shops: ShopsState;
  ui: UiState;
}

export const reducers: ActionReducerMap<AppState> = {
  auth: authReducer,
  shops: shopsReducer,
  ui: uiReducer,
};
