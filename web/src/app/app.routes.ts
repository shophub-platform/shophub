import { Routes } from '@angular/router';
import { authGuard } from './core/guards/auth.guard';

export const routes: Routes = [
  // Javne rute
  {
    path: 'login',
    title: 'ShopHub — Prijava',
    loadComponent: () =>
      import('./features/auth/login.component').then((m) => m.LoginComponent),
  },
  {
    path: 'register',
    title: 'ShopHub — Registracija',
    loadComponent: () =>
      import('./features/auth/register.component').then((m) => m.RegisterComponent),
  },

  // Zaštićene rute (zahtevaju JWT)
  {
    path: '',
    canActivate: [authGuard],
    title: 'ShopHub — Dashboard',
    loadComponent: () =>
      import('./features/dashboard/dashboard.component').then((m) => m.DashboardComponent),
  },
  {
    path: 'shops/new',
    canActivate: [authGuard],
    title: 'ShopHub — Nova prodavnica',
    loadComponent: () =>
      import('./features/shops/shop-wizard.component').then((m) => m.ShopWizardComponent),
  },
  {
    path: 'shops/:id',
    canActivate: [authGuard],
    title: 'ShopHub — Detalji prodavnice',
    loadComponent: () =>
      import('./features/shops/shop-detail.component').then((m) => m.ShopDetailComponent),
  },

  { path: '**', redirectTo: '' },
];
