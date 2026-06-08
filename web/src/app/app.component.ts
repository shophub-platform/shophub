import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink, RouterOutlet, RouterLinkActive } from '@angular/router';
import { Store } from '@ngrx/store';

import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';

import { combineLatest, map } from 'rxjs';

import { selectAuthenticated, selectAuthLoading } from './store/auth/auth.selectors';
import { selectGlobalLoading } from './store/ui/ui.selectors';
import {
  selectShopsLoading,
  selectShopsSaving,
} from './store/shops/shops.selectors';
import { AuthActions } from './store/auth/auth.actions';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [
    CommonModule,
    RouterOutlet,
    RouterLink,
    RouterLinkActive,
    MatToolbarModule,
    MatButtonModule,
    MatIconModule,
    MatProgressBarModule,
  ],
  template: `
    <mat-toolbar color="primary" class="sh-row">
      <mat-icon>storefront</mat-icon>
      <span style="font-weight:600;margin-left:8px">ShopHub</span>

      @if (authenticated$ | async) {
        <nav class="sh-row" style="margin-left:24px;gap:8px">
          <a mat-button routerLink="/" routerLinkActive="active-link" [routerLinkActiveOptions]="{ exact: true }">Dashboard</a>
          <a mat-button routerLink="/shops/new" routerLinkActive="active-link">New shop</a>
        </nav>
      }

      <span class="sh-spacer"></span>

      @if (authenticated$ | async) {
        <button mat-button (click)="logout()">
          <mat-icon>logout</mat-icon> Sign out
        </button>
      }
    </mat-toolbar>

    @if (loading$ | async) {
      <mat-progress-bar mode="indeterminate" aria-label="Loading"></mat-progress-bar>
    }

    <main>
      <router-outlet />
    </main>

    <footer class="sh-footer">
      <div class="sh-footer-inner">
        <div class="sh-footer-brand">
          <mat-icon>storefront</mat-icon>
          <span>ShopHub</span>
          <span class="sh-muted">— Kubernetes-native shop platform</span>
        </div>

        <nav class="sh-footer-links">
          <a href="#">Docs</a>
          <a href="#">API</a>
          <a href="#">Status</a>
          <a href="https://github.com/shophub-platform" target="_blank" rel="noopener">GitHub</a>
        </nav>

        <div class="sh-footer-badges">
          <span class="badge">Angular</span>
          <span class="badge">NgRx</span>
          <span class="badge">Material</span>
          <span class="badge">Go</span>
          <span class="badge">Kubernetes</span>
        </div>
      </div>
      <div class="sh-footer-bottom sh-muted">© {{ year }} ShopHub · All rights reserved.</div>
    </footer>
  `,
  styles: [
    `
      :host {
        display: flex;
        flex-direction: column;
        min-height: 100vh;
      }
      main {
        flex: 1 0 auto;
      }
      .active-link {
        background: rgba(255, 255, 255, 0.16);
        border-radius: 6px;
      }
      mat-toolbar mat-icon {
        vertical-align: middle;
      }

      .sh-footer {
        flex-shrink: 0;
        margin-top: 32px;
        background: #161922;
        color: #cbd2e0;
      }
      .sh-footer-inner {
        max-width: 1080px;
        margin: 0 auto;
        padding: 28px 16px 16px;
        display: flex;
        flex-wrap: wrap;
        gap: 16px 32px;
        align-items: center;
      }
      .sh-footer-brand {
        display: flex;
        align-items: center;
        gap: 8px;
        font-weight: 600;
        font-size: 16px;
      }
      .sh-footer-brand .sh-muted {
        color: #8b93a7;
        font-weight: 400;
        font-size: 13px;
      }
      .sh-footer-links {
        display: flex;
        gap: 18px;
        margin-left: auto;
      }
      .sh-footer-links a {
        color: #cbd2e0;
        text-decoration: none;
        font-size: 14px;
        transition: color 0.2s ease;
      }
      .sh-footer-links a:hover {
        color: #fff;
      }
      .sh-footer-badges {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;
        width: 100%;
      }
      .badge {
        font-size: 12px;
        padding: 3px 10px;
        border-radius: 999px;
        background: rgba(255, 255, 255, 0.08);
        border: 1px solid rgba(255, 255, 255, 0.12);
        transition: transform 0.2s ease, background 0.2s ease;
      }
      .badge:hover {
        transform: translateY(-2px);
        background: rgba(124, 58, 237, 0.3);
      }
      .sh-footer-bottom {
        text-align: center;
        padding: 14px 16px 20px;
        font-size: 12.5px;
        color: #6e7689 !important;
        border-top: 1px solid rgba(255, 255, 255, 0.06);
      }
    `,
  ],
})
export class AppComponent {
  readonly year = new Date().getFullYear();
  private readonly store = inject(Store);
  readonly authenticated$ = this.store.select(selectAuthenticated);

  // Globalna traka napretka: aktivna kada se bilo gde čita/snima (FZ 4, 9.1).
  readonly loading$ = combineLatest([
    this.store.select(selectGlobalLoading),
    this.store.select(selectAuthLoading),
    this.store.select(selectShopsLoading),
    this.store.select(selectShopsSaving),
  ]).pipe(map((flags) => flags.some(Boolean)));

  logout(): void {
    this.store.dispatch(AuthActions.logout());
  }
}
