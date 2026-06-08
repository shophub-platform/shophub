import { createFeatureSelector, createSelector } from '@ngrx/store';
import { AuthState } from './auth.reducer';

export const selectAuth = createFeatureSelector<AuthState>('auth');

export const selectAuthenticated = createSelector(selectAuth, (s) => s.authenticated);
export const selectAuthLoading = createSelector(selectAuth, (s) => s.loading);
export const selectAuthError = createSelector(selectAuth, (s) => s.error);
export const selectCurrentUser = createSelector(selectAuth, (s) => s.user);
