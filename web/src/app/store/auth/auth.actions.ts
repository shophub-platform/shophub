import { createActionGroup, emptyProps, props } from '@ngrx/store';
import {
  Credentials,
  RegisterPayload,
  TokenResponse,
  CurrentUser,
} from '../../core/models/auth.model';

export const AuthActions = createActionGroup({
  source: 'Auth',
  events: {
    Login: props<{ creds: Credentials; redirect?: string }>(),
    'Login Success': props<{ tokens: TokenResponse; redirect?: string }>(),
    'Login Failure': props<{ error: string }>(),

    Register: props<{ payload: RegisterPayload }>(),
    'Register Success': props<{ tokens: TokenResponse }>(),
    'Register Failure': props<{ error: string }>(),

    'Load Me': emptyProps(),
    'Load Me Success': props<{ user: CurrentUser }>(),

    Logout: emptyProps(),
  },
});
