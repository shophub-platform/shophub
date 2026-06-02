import { Component } from '@angular/core';
import { RouterLink, RouterOutlet } from '@angular/router';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, RouterLink],
  template: `
    <header style="padding:12px 20px;border-bottom:1px solid #e5e7eb;display:flex;gap:16px;align-items:center">
      <strong>ShopHub</strong>
      <nav style="display:flex;gap:12px">
        <a routerLink="/">Dashboard</a>
        <a routerLink="/shops">Shops</a>
      </nav>
    </header>
    <main style="padding:20px">
      <router-outlet />
    </main>
  `,
})
export class AppComponent {}
