import { createReducer, on } from '@ngrx/store';
import { UiActions } from './ui.actions';

export interface UiState {
  loading: boolean;
}

export const initialUiState: UiState = {
  loading: false,
};

export const uiReducer = createReducer(
  initialUiState,
  on(UiActions.setLoading, (state, { loading }) => ({ ...state, loading })),
);
