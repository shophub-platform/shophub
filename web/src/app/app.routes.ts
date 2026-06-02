import { Routes } from '@angular/router';
import { DashboardComponent } from './pages/dashboard/dashboard.component';

export const routes: Routes = [
  // Eager-loaded ruta
  { path: '', component: DashboardComponent, title: 'ShopHub — Dashboard' },

  // Lazy-loaded feature modul (prodavnice)
  {
    path: 'shops',
    loadChildren: () =>
      import('./pages/shops/shops.routes').then((m) => m.SHOPS_ROUTES),
  },

  { path: '**', redirectTo: '' },
];
