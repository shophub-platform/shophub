import { createFeatureSelector, createSelector } from '@ngrx/store';
import { UiState } from './ui.reducer';

export const selectUi = createFeatureSelector<UiState>('ui');
export const selectGlobalLoading = createSelector(selectUi, (s) => s.loading);
