import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { Store } from '@ngrx/store';

import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';

import { AuthActions } from '../../store/auth/auth.actions';
import { selectAuthError, selectAuthLoading } from '../../store/auth/auth.selectors';

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
  ],
  template: `
    <div class="auth-page">
      <mat-card style="max-width:430px;width:100%">
        <div class="auth-brand">
          <div class="auth-logo"><mat-icon>storefront</mat-icon></div>
          <h2>Create your account</h2>
          <p class="sh-muted">Launch your shops on ShopHub</p>
        </div>

        <mat-card-content>
          <form [formGroup]="form" (ngSubmit)="submit()" class="full-width">
            <mat-form-field appearance="outline" class="full-width">
              <mat-label>Display name</mat-label>
              <mat-icon matPrefix>person</mat-icon>
              <input matInput formControlName="displayName" autocomplete="name" placeholder="Jane Doe" />
              @if (form.controls.displayName.touched && form.controls.displayName.invalid) {
                <mat-error>Name is required.</mat-error>
              }
            </mat-form-field>

            <mat-form-field appearance="outline" class="full-width">
              <mat-label>Email</mat-label>
              <mat-icon matPrefix>mail</mat-icon>
              <input matInput type="email" formControlName="email" autocomplete="username" placeholder="you@example.com" />
              @if (form.controls.email.touched && form.controls.email.invalid) {
                <mat-error>Enter a valid email.</mat-error>
              }
            </mat-form-field>

            <mat-form-field appearance="outline" class="full-width">
              <mat-label>Password</mat-label>
              <mat-icon matPrefix>lock</mat-icon>
              <input matInput type="password" formControlName="password" autocomplete="new-password" />
              <mat-hint>At least 8 characters.</mat-hint>
              @if (form.controls.password.touched && form.controls.password.invalid) {
                <mat-error>Password must be at least 8 characters.</mat-error>
              }
            </mat-form-field>

            @if (error$ | async; as error) {
              <p style="color:var(--sh-status-error);margin:4px 0">{{ error }}</p>
            }

            <button
              mat-flat-button
              color="primary"
              class="full-width"
              type="submit"
              [disabled]="(loading$ | async) || form.invalid"
            >
              @if (loading$ | async) {
                <mat-progress-spinner diameter="20" mode="indeterminate"></mat-progress-spinner>
              } @else {
                <mat-icon>person_add</mat-icon> Create account
              }
            </button>
          </form>
        </mat-card-content>

        <mat-card-actions align="end">
          <span class="sh-muted">Already have an account?</span>
          <a mat-button color="primary" routerLink="/login">Sign in</a>
        </mat-card-actions>
      </mat-card>
    </div>
  `,
})
export class RegisterComponent {
  private readonly fb = inject(FormBuilder);
  private readonly store = inject(Store);

  readonly loading$ = this.store.select(selectAuthLoading);
  readonly error$ = this.store.select(selectAuthError);

  readonly form = this.fb.nonNullable.group({
    displayName: ['', [Validators.required]],
    email: ['', [Validators.required, Validators.email]],
    password: ['', [Validators.required, Validators.minLength(8)]],
  });

  submit(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    this.store.dispatch(AuthActions.register({ payload: this.form.getRawValue() }));
  }
}
