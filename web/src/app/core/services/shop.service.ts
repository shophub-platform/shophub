import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import {
  Shop,
  CreateShopPayload,
  UpdateShopPayload,
} from '../models/shop.model';

/** ShopService barata sa /api/v1/shops (FZ 2). */
@Injectable({ providedIn: 'root' })
export class ShopService {
  private readonly http = inject(HttpClient);
  private readonly base = '/api/v1/shops';

  list(): Observable<Shop[]> {
    return this.http.get<Shop[]>(this.base);
  }

  get(id: string): Observable<Shop> {
    return this.http.get<Shop>(`${this.base}/${id}`);
  }

  create(payload: CreateShopPayload): Observable<Shop> {
    return this.http.post<Shop>(this.base, payload);
  }

  update(id: string, patch: UpdateShopPayload): Observable<Shop> {
    return this.http.patch<Shop>(`${this.base}/${id}`, patch);
  }

  remove(id: string): Observable<void> {
    return this.http.delete<void>(`${this.base}/${id}`);
  }

  getUrl(id: string): Observable<{ url: string }> {
    return this.http.get<{ url: string }>(`${this.base}/${id}/url`);
  }
}
