import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap } from 'rxjs';

import {
  Credentials,
  RegisterPayload,
  TokenResponse,
  CurrentUser,
} from '../models/auth.model';

const ACCESS_KEY = 'sh_access';
const REFRESH_KEY = 'sh_refresh';

/** AuthService barata sa /api/v1/auth/* i čuva tokene u localStorage. */
@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly base = '/api/v1/auth';

  register(payload: RegisterPayload): Observable<TokenResponse> {
    return this.http
      .post<TokenResponse>(`${this.base}/register`, payload)
      .pipe(tap((t) => this.saveTokens(t)));
  }

  login(creds: Credentials): Observable<TokenResponse> {
    return this.http
      .post<TokenResponse>(`${this.base}/login`, creds)
      .pipe(tap((t) => this.saveTokens(t)));
  }

  refresh(): Observable<TokenResponse> {
    return this.http
      .post<TokenResponse>(`${this.base}/refresh`, {
        refreshToken: this.getRefreshToken(),
      })
      .pipe(tap((t) => this.saveTokens(t)));
  }

  me(): Observable<CurrentUser> {
    return this.http.get<CurrentUser>('/api/v1/me');
  }

  logout(): void {
    localStorage.removeItem(ACCESS_KEY);
    localStorage.removeItem(REFRESH_KEY);
  }

  saveTokens(t: TokenResponse): void {
    localStorage.setItem(ACCESS_KEY, t.accessToken);
    localStorage.setItem(REFRESH_KEY, t.refreshToken);
  }

  getAccessToken(): string | null {
    return localStorage.getItem(ACCESS_KEY);
  }

  getRefreshToken(): string | null {
    return localStorage.getItem(REFRESH_KEY);
  }

  isAuthenticated(): boolean {
    return !!this.getAccessToken();
  }
}
