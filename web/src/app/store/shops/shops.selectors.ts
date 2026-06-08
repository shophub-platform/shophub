import { createFeatureSelector, createSelector } from '@ngrx/store';
import { ShopsState } from './shops.reducer';

export const selectShopsState = createFeatureSelector<ShopsState>('shops');

export const selectShops = createSelector(selectShopsState, (s) => s.list);
export const selectShopsLoading = createSelector(selectShopsState, (s) => s.loading);
export const selectShopsSaving = createSelector(selectShopsState, (s) => s.saving);
export const selectSelectedShop = createSelector(selectShopsState, (s) => s.selected);
export const selectShopsError = createSelector(selectShopsState, (s) => s.error);

export const selectShopsCount = createSelector(selectShops, (list) => list.length);
