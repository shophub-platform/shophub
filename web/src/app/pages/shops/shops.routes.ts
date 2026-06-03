import { Routes } from '@angular/router';
import { ShopsComponent } from './shops.component';

// Lazy-loaded rute za feature "prodavnice".
export const SHOPS_ROUTES: Routes = [
  { path: '', component: ShopsComponent, title: 'ShopHub — Prodavnice' },
];
