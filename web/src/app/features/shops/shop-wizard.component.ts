import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Store } from '@ngrx/store';

import { MatCardModule } from '@angular/material/card';
import { MatStepperModule } from '@angular/material/stepper';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';

import { ShopsActions } from '../../store/shops/shops.actions';
import { selectShopsSaving } from '../../store/shops/shops.selectors';
import { CreateShopPayload } from '../../core/models/shop.model';

@Component({
  selector: 'app-shop-wizard',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatCardModule,
    MatStepperModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatButtonModule,
    MatIconModule,
  ],
  template: `
    <div class="sh-container" style="max-width:720px">
      <h1 style="display:flex;align-items:center;gap:10px;margin-bottom:4px">
        <mat-icon style="color:var(--sh-brand-1)">add_business</mat-icon> New shop
      </h1>
      <p class="sh-muted" style="margin:0 0 20px">
        Complete three steps and ShopHub will provision your shop.
      </p>

      <mat-card>
        <mat-card-content>
          <mat-stepper orientation="vertical" linear #stepper>
            <!-- Korak 1: Osnovno -->
            <mat-step [stepControl]="basic">
              <ng-template matStepLabel>Basics — name &amp; availability</ng-template>
              <form [formGroup]="basic">
                <mat-form-field appearance="outline" class="full-width">
                  <mat-label>Shop name</mat-label>
                  <input matInput formControlName="name" placeholder="my-store" />
                  @if (basic.controls.name.touched && basic.controls.name.invalid) {
                    <mat-error>Name is required.</mat-error>
                  }
                </mat-form-field>

                <mat-form-field appearance="outline" class="full-width">
                  <mat-label>Availability</mat-label>
                  <mat-select formControlName="availability">
                    <mat-option value="standard">Standard (2 replicas)</mat-option>
                    <mat-option value="high">High (3 replicas)</mat-option>
                  </mat-select>
                </mat-form-field>

                <mat-form-field appearance="outline" class="full-width">
                  <mat-label>Container image (optional)</mat-label>
                  <input matInput formControlName="image" placeholder="ghcr.io/.../shop:latest" />
                </mat-form-field>

                <div align="end">
                  <button mat-flat-button color="primary" matStepperNext [disabled]="basic.invalid">
                    Next
                  </button>
                </div>
              </form>
            </mat-step>

            <!-- Korak 2: Wallet -->
            <mat-step [stepControl]="wallet">
              <ng-template matStepLabel>Wallet — payout address</ng-template>
              <form [formGroup]="wallet">
                <mat-form-field appearance="outline" class="full-width">
                  <mat-label>Wallet address</mat-label>
                  <input matInput formControlName="walletAddress" placeholder="0x..." />
                  <mat-hint>Address that receives payments.</mat-hint>
                  @if (wallet.controls.walletAddress.touched && wallet.controls.walletAddress.invalid) {
                    <mat-error>Wallet address is required.</mat-error>
                  }
                </mat-form-field>

                <div align="end">
                  <button mat-button matStepperPrevious>Back</button>
                  <button mat-flat-button color="primary" matStepperNext [disabled]="wallet.invalid">
                    Next
                  </button>
                </div>
              </form>
            </mat-step>

            <!-- Korak 3: Baza -->
            <mat-step [stepControl]="database">
              <ng-template matStepLabel>Database — backend choice</ng-template>
              <form [formGroup]="database">
                <mat-form-field appearance="outline" class="full-width">
                  <mat-label>Database type</mat-label>
                  <mat-select formControlName="databaseType">
                    <mat-option value="postgres">PostgreSQL</mat-option>
                    <mat-option value="redis">Redis</mat-option>
                  </mat-select>
                </mat-form-field>

                <div align="end">
                  <button mat-button matStepperPrevious>Back</button>
                  <button
                    mat-flat-button
                    color="primary"
                    (click)="create()"
                    [disabled]="saving$ | async"
                  >
                    <mat-icon>rocket_launch</mat-icon> Create shop
                  </button>
                </div>
              </form>
            </mat-step>
          </mat-stepper>
        </mat-card-content>
      </mat-card>
    </div>
  `,
})
export class ShopWizardComponent {
  private readonly fb = inject(FormBuilder);
  private readonly store = inject(Store);

  readonly saving$ = this.store.select(selectShopsSaving);

  readonly basic = this.fb.nonNullable.group({
    name: ['', [Validators.required, Validators.maxLength(63)]],
    availability: this.fb.nonNullable.control<'standard' | 'high'>('standard'),
    image: [''],
  });

  readonly wallet = this.fb.nonNullable.group({
    walletAddress: ['', [Validators.required]],
  });

  readonly database = this.fb.nonNullable.group({
    databaseType: this.fb.nonNullable.control<'postgres' | 'redis'>('postgres'),
  });

  create(): void {
    if (this.basic.invalid || this.wallet.invalid || this.database.invalid) {
      return;
    }
    const b = this.basic.getRawValue();
    const payload: CreateShopPayload = {
      name: b.name,
      availability: b.availability,
      image: b.image || undefined,
      walletAddress: this.wallet.getRawValue().walletAddress,
      databaseType: this.database.getRawValue().databaseType,
    };
    this.store.dispatch(ShopsActions.create({ payload }));
  }
}
