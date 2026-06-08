import { createReducer, on } from '@ngrx/store';
import { AuthActions } from './auth.actions';
import { CurrentUser } from '../../core/models/auth.model';

export interface AuthState {
  authenticated: boolean;
  user: CurrentUser | null;
  loading: boolean;
  error: string | null;
}

export const initialAuthState: AuthState = {
  // Inicijalno čitamo iz localStorage da preživi refresh stranice.
  authenticated: !!localStorage.getItem('sh_access'),
  user: null,
  loading: false,
  error: null,
};

export const authReducer = createReducer(
  initialAuthState,
  on(AuthActions.login, AuthActions.register, (s) => ({
    ...s,
    loading: true,
    error: null,
  })),
  on(AuthActions.loginSuccess, AuthActions.registerSuccess, (s) => ({
    ...s,
    loading: false,
    authenticated: true,
    error: null,
  })),
  on(AuthActions.loginFailure, AuthActions.registerFailure, (s, { error }) => ({
    ...s,
    loading: false,
    error,
  })),
  on(AuthActions.loadMeSuccess, (s, { user }) => ({ ...s, user })),
  on(AuthActions.logout, (s) => ({
    ...s,
    authenticated: false,
    user: null,
    error: null,
  })),
);
