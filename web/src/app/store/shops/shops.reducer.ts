import { createReducer, on } from '@ngrx/store';
import { ShopsActions } from './shops.actions';
import { Shop } from '../../core/models/shop.model';

export interface ShopsState {
  list: Shop[];
  selected: Shop | null;
  loading: boolean;
  saving: boolean;
  error: string | null;
}

export const initialShopsState: ShopsState = {
  list: [],
  selected: null,
  loading: false,
  saving: false,
  error: null,
};

export const shopsReducer = createReducer(
  initialShopsState,

  on(ShopsActions.load, (s) => ({ ...s, loading: true, error: null })),
  on(ShopsActions.loadSuccess, (s, { shops }) => ({
    ...s,
    loading: false,
    list: shops,
  })),
  on(ShopsActions.loadFailure, (s, { error }) => ({
    ...s,
    loading: false,
    error,
  })),

  on(ShopsActions.loadOne, (s) => ({ ...s, loading: true, selected: null, error: null })),
  on(ShopsActions.loadOneSuccess, (s, { shop }) => ({
    ...s,
    loading: false,
    selected: shop,
  })),
  on(ShopsActions.loadOneFailure, (s, { error }) => ({
    ...s,
    loading: false,
    error,
  })),

  on(ShopsActions.create, ShopsActions.update, (s) => ({
    ...s,
    saving: true,
    error: null,
  })),
  on(ShopsActions.createSuccess, (s, { shop }) => ({
    ...s,
    saving: false,
    list: [shop, ...s.list],
    selected: shop,
  })),
  on(ShopsActions.updateSuccess, (s, { shop }) => ({
    ...s,
    saving: false,
    selected: shop,
    list: s.list.map((x) => (x.id === shop.id ? shop : x)),
  })),
  on(ShopsActions.createFailure, ShopsActions.updateFailure, (s, { error }) => ({
    ...s,
    saving: false,
    error,
  })),

  on(ShopsActions.delete, (s) => ({ ...s, saving: true })),
  on(ShopsActions.deleteSuccess, (s, { id }) => ({
    ...s,
    saving: false,
    list: s.list.filter((x) => x.id !== id),
    selected: s.selected?.id === id ? null : s.selected,
  })),
  on(ShopsActions.deleteFailure, (s, { error }) => ({
    ...s,
    saving: false,
    error,
  })),
);
