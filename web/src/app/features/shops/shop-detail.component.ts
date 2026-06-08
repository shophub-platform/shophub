import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute } from '@angular/router';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Store } from '@ngrx/store';

import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { filter } from 'rxjs';

import { StatusChipComponent } from '../../shared/status-chip.component';
import {
  ConfirmDialogComponent,
  ConfirmData,
} from '../../shared/confirm-dialog.component';
import { ShopsActions } from '../../store/shops/shops.actions';
import {
  selectSelectedShop,
  selectShopsLoading,
  selectShopsSaving,
} from '../../store/shops/shops.selectors';

@Component({
  selector: 'app-shop-detail',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
    MatDialogModule,
    StatusChipComponent,
  ],
  template: `
    <div class="sh-container" style="max-width:720px">
      @if (loading$ | async) {
        <div class="sh-center"><mat-progress-spinner mode="indeterminate" diameter="48" /></div>
      } @else {
        @if (shop$ | async; as shop) {
          <div class="sh-row" style="margin-bottom:8px">
            <h1 style="margin:0">{{ shop.name }}</h1>
            <app-status-chip [status]="shop.status" />
            <span class="sh-spacer"></span>
            <a mat-stroked-button [href]="shop.url" target="_blank" rel="noopener">
              <mat-icon>open_in_new</mat-icon> Open shop
            </a>
          </div>

          <mat-card style="margin-top:12px">
            <mat-card-header>
              <mat-card-title style="display:flex;align-items:center;gap:8px">
                <mat-icon style="color:var(--sh-brand-1)">tune</mat-icon> Configuration
              </mat-card-title>
            </mat-card-header>
            <mat-card-content>
              <form [formGroup]="form" (ngSubmit)="save(shop.id)">
                <mat-form-field appearance="outline" class="full-width">
                  <mat-label>Availability</mat-label>
                  <mat-select formControlName="availability">
                    <mat-option value="standard">Standard (2 replicas)</mat-option>
                    <mat-option value="high">High (3 replicas)</mat-option>
                  </mat-select>
                </mat-form-field>

                <mat-form-field appearance="outline" class="full-width">
                  <mat-label>Database type</mat-label>
                  <mat-select formControlName="databaseType">
                    <mat-option value="postgres">PostgreSQL</mat-option>
                    <mat-option value="redis">Redis</mat-option>
                  </mat-select>
                </mat-form-field>

                <mat-form-field appearance="outline" class="full-width">
                  <mat-label>Wallet address</mat-label>
                  <input matInput formControlName="walletAddress" />
                  @if (form.controls.walletAddress.touched && form.controls.walletAddress.invalid) {
                    <mat-error>Wallet address is required.</mat-error>
                  }
                </mat-form-field>

                <div class="sh-row">
                  <button
                    mat-flat-button
                    color="primary"
                    type="submit"
                    [disabled]="(saving$ | async) || form.invalid || form.pristine"
                  >
                    Save changes
                  </button>
                  <span class="sh-spacer"></span>
                  <button mat-stroked-button color="warn" type="button" (click)="confirmDelete(shop.id, shop.name)">
                    <mat-icon>delete</mat-icon> Delete
                  </button>
                </div>
              </form>
            </mat-card-content>
          </mat-card>

          <p class="sh-muted" style="margin-top:12px">
            ID: {{ shop.id }} · Created: {{ shop.createdAt | date: 'short' }} · URL: {{ shop.url }}
          </p>
        } @else {
          <mat-card><mat-card-content>Shop not found.</mat-card-content></mat-card>
        }
      }
    </div>
  `,
})
export class ShopDetailComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly store = inject(Store);
  private readonly route = inject(ActivatedRoute);
  private readonly dialog = inject(MatDialog);

  readonly shop$ = this.store.select(selectSelectedShop);
  readonly loading$ = this.store.select(selectShopsLoading);
  readonly saving$ = this.store.select(selectShopsSaving);

  readonly form = this.fb.nonNullable.group({
    availability: this.fb.nonNullable.control<'standard' | 'high'>('standard'),
    databaseType: this.fb.nonNullable.control<'postgres' | 'redis'>('postgres'),
    walletAddress: ['', [Validators.required]],
  });

  constructor() {
    // Popuni formu kad stigne izabrana prodavnica.
    this.shop$
      .pipe(
        filter((s) => !!s),
        takeUntilDestroyed(),
      )
      .subscribe((s) => {
        this.form.reset({
          availability: s!.availability,
          databaseType: s!.databaseType,
          walletAddress: s!.walletAddress,
        });
      });
  }

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.store.dispatch(ShopsActions.loadOne({ id }));
    }
  }

  save(id: string): void {
    if (this.form.invalid) {
      return;
    }
    // Šaljemo samo izmenjena polja (PATCH).
    this.store.dispatch(ShopsActions.update({ id, patch: this.form.getRawValue() }));
    this.form.markAsPristine();
  }

  confirmDelete(id: string, name: string): void {
    const data: ConfirmData = {
      title: 'Delete shop',
      message: `Are you sure you want to delete "${name}"? The operator will remove all resources.`,
      confirmText: 'Delete',
      danger: true,
    };
    this.dialog
      .open(ConfirmDialogComponent, { data, width: '420px' })
      .afterClosed()
      .subscribe((ok) => {
        if (ok) {
          this.store.dispatch(ShopsActions.delete({ id, name }));
        }
      });
  }
}
