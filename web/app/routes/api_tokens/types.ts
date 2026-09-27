// Mirrors handle_api_tokens.go's JSON shape exactly, snake_case included
// (§3.5 of docs/spa-migration.md).

export type APITokenScope = "read" | "write";

export interface APIToken {
  id: number;
  name: string;
  /** The token masked to its last four characters: `nin_••••wTEF`. */
  masked: string;
  scope: APITokenScope;
  /** Unix seconds, or null for a token that never expires. */
  expires_at: number | null;
  last_used_at: number | null;
  created_at: number;
  /** Computed by the server, so the page never compares against its own clock. */
  expired: boolean;
}

export interface APITokensResponse {
  data: APIToken[];
  /** How many active tokens the server allows (logic.APITokenLimit). */
  limit: number;
  /** The lifetimes the server accepts, in days; 0 means never expires. */
  expiry_days: number[];
}

export interface APITokenCreatedResponse {
  token: APIToken;
  /** The plaintext token. Sent once, at creation, and never again. */
  secret: string;
}
