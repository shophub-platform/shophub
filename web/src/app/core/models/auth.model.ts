// Auth modeli — usklađeni sa ShopHub backend-om (FZ 2: /api/v1/auth/*).

export interface Credentials {
  email: string;
  password: string;
}

export interface RegisterPayload extends Credentials {
  displayName: string;
}

export interface TokenResponse {
  accessToken: string;
  refreshToken: string;
  tokenType: string;
}

export interface CurrentUser {
  userID: string;
  email: string;
}
