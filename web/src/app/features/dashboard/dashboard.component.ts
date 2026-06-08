import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { Store } from '@ngrx/store';

import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';

import { StatusChipComponent } from '../../shared/status-chip.component';
import { HeroCarouselComponent } from '../../shared/hero-carousel.component';
import { ShopsActions } from '../../store/shops/shops.actions';
import {
  selectShops,
  selectShopsLoading,
} from '../../store/shops/shops.selectors';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
    StatusChipComponent,
    HeroCarouselComponent,
  ],
  template: `
    <div class="sh-container">
      <app-hero-carousel />

      <div class="sh-row" style="margin-bottom:20px">
        <div>
          <h1 style="margin:0;display:flex;align-items:center;gap:10px">
            <mat-icon style="color:var(--sh-brand-1)">storefront</mat-icon> My shops
          </h1>
          <p class="sh-muted" style="margin:4px 0 0">
            Manage all your shop sites in one place.
          </p>
        </div>
        <span class="sh-spacer"></span>
        <a mat-flat-button color="primary" routerLink="/shops/new">
          <mat-icon>add_business</mat-icon> New shop
        </a>
      </div>

      @if (loading$ | async) {
        <div class="sh-center"><mat-progress-spinner mode="indeterminate" diameter="48" /></div>
      } @else {
        @if (shops$ | async; as shops) {
          @if (shops.length === 0) {
            <mat-card>
              <mat-card-content style="text-align:center;padding:48px">
                <mat-icon style="font-size:48px;width:48px;height:48px;color:var(--sh-muted)">storefront</mat-icon>
                <p class="sh-muted">You don't have any shops yet.</p>
                <a mat-flat-button color="primary" routerLink="/shops/new">Create your first</a>
              </mat-card-content>
            </mat-card>
          } @else {
            <div class="sh-grid">
              @for (shop of shops; track shop.id) {
                <mat-card class="shop-card">
                  <mat-card-header>
                    <mat-card-title>{{ shop.name }}</mat-card-title>
                    <mat-card-subtitle>
                      <app-status-chip [status]="shop.status" />
                    </mat-card-subtitle>
                  </mat-card-header>
                  <mat-card-content>
                    <div class="card-meta">
                      <span><mat-icon class="inline-icon">dns</mat-icon> {{ shop.databaseType }}</span>
                      <span><mat-icon class="inline-icon">layers</mat-icon> {{ shop.availability }}</span>
                    </div>
                    <div class="card-meta">
                      <span><mat-icon class="inline-icon">account_balance_wallet</mat-icon> {{ shop.walletAddress | slice: 0 : 12 }}…</span>
                    </div>
                    <div class="card-meta">
                      <span><mat-icon class="inline-icon">schedule</mat-icon> {{ shop.createdAt | date: 'dd.MM.yyyy.' }}</span>
                    </div>
                  </mat-card-content>
                  <mat-card-actions>
                    <a mat-button color="primary" [routerLink]="['/shops', shop.id]">Details</a>
                    <a mat-button [href]="shop.url" target="_blank" rel="noopener">
                      <mat-icon>open_in_new</mat-icon> Open shop
                    </a>
                  </mat-card-actions>
                </mat-card>
              }
            </div>
          }
        }
      }
    </div>
  `,
  styles: [
    `
      .inline-icon {
        font-size: 16px;
        width: 16px;
        height: 16px;
        vertical-align: text-bottom;
        color: var(--sh-muted);
      }
      .card-meta {
        display: flex;
        gap: 16px;
        color: var(--sh-muted);
        font-size: 13px;
        margin: 6px 0 0;
      }
      .card-meta span {
        display: inline-flex;
        align-items: center;
        gap: 5px;
      }
      .shop-card {
        transition: box-shadow 0.15s ease;
      }
      .shop-card:hover {
        box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
      }
    `,
  ],
})
export class DashboardComponent implements OnInit {
  private readonly store = inject(Store);
  readonly shops$ = this.store.select(selectShops);
  readonly loading$ = this.store.select(selectShopsLoading);

  ngOnInit(): void {
    this.store.dispatch(ShopsActions.load());
  }
}
