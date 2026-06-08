// Shop modeli — usklađeni sa shopResponse iz backend-a (FZ 2: /api/v1/shops).

export type Availability = 'standard' | 'high';
export type DatabaseType = 'postgres' | 'redis';

// Status koji vraća backend (izveden iz Shop CR .status.phase).
export type ShopStatus = 'pending' | 'deploying' | 'running' | 'error';

export interface Shop {
  id: string;
  name: string;
  availability: Availability;
  walletAddress: string;
  databaseType: DatabaseType;
  image: string;
  status: ShopStatus;
  url: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateShopPayload {
  name: string;
  availability: Availability;
  walletAddress: string;
  databaseType: DatabaseType;
  image?: string;
}

// PATCH — sva polja opciona (delimična izmena konfiguracije).
export interface UpdateShopPayload {
  availability?: Availability;
  walletAddress?: string;
  databaseType?: DatabaseType;
}

// Prikazna oznaka i CSS klasa za status (Provisioning / Ready / Error...).
export const STATUS_META: Record<ShopStatus, { label: string; css: string }> = {
  pending: { label: 'Pending', css: 'status-pending' },
  deploying: { label: 'Provisioning', css: 'status-deploying' },
  running: { label: 'Ready', css: 'status-running' },
  error: { label: 'Error', css: 'status-error' },
};
