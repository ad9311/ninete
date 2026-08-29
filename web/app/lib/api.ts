// The single place the SPA talks to /api/*. Everything here follows §3.1, §3.2
// and §3.5 of docs/spa-migration.md; the comments record the parts that are
// invisible in the code but expensive to rediscover.

import { begin, end } from "./pending";

const API_PREFIX = "/api";
// Points into the SPA (Phase 6 of docs/spa-migration.md, settling the
// TODO.md note on where session expiry redirects). Go redirects to the same
// place through handlers.AppLoginPath, which a bundle cannot import — the two
// literals have to be changed together.
const LOGIN_PATH = "/login";

/** The error envelope every /api/* failure carries (handlers.APIError). */
interface APIErrorBody {
  error?: string;
  fields?: Record<string, string>;
}

/**
 * A non-2xx answer from the API. `fields` is filled only by a 422, and its keys
 * are the snake_case field names the request itself uses (§3.5) — the value is
 * the validator's rule (`required`, `gte`), not a sentence, so the phrasing is
 * the caller's job.
 */
export class APIRequestError extends Error {
  readonly status: number;
  readonly fields: Record<string, string>;

  constructor(status: number, message: string, fields: Record<string, string>) {
    super(message);
    this.name = "APIRequestError";
    this.status = status;
    this.fields = fields;
  }

  /** True when the server rejected the payload rather than failing. */
  get isValidation(): boolean {
    return this.status === 422;
  }
}

// nosurf sets ninete_csrf as HttpOnly, so the token cannot be read from the
// cookie — the shell prints it into a <meta> tag instead. It survives login and
// logout (nosurf regenerates only when its own cookie is missing), so reading
// it once per page load is correct and a session boundary does not invalidate
// the cached value.
let cachedCSRFToken: string | null = null;

export function csrfToken(): string {
  if (cachedCSRFToken === null) {
    const meta = document.querySelector<HTMLMetaElement>(
      'meta[name="csrf-token"]',
    );
    cachedCSRFToken = meta?.content ?? "";
  }

  return cachedCSRFToken;
}

/** Test seam: drops the memoized token so a re-rendered shell is picked up. */
export function resetCSRFToken(): void {
  cachedCSRFToken = null;
}

export interface RequestOptions {
  /** Query parameters appended to the path. Undefined values are dropped. */
  params?: Record<string, string | number | boolean | undefined>;
  signal?: AbortSignal;
  /**
   * Skip the global redirect-to-login on a 401. Header.svelte's own session
   * probe sets this: it is the one /api/* call every route makes, including
   * the guest-reachable /login and /register (Phase 6), and without this a
   * guest landing on either would be bounced straight back to LOGIN_PATH by
   * their own session check.
   */
  skipAuthRedirect?: boolean;
}

function buildURL(path: string, params?: RequestOptions["params"]): string {
  // Match the prefix as a whole path segment. A plain startsWith would treat a
  // future "/api-tokens" as already prefixed and fetch it unprefixed.
  const prefixed = path === API_PREFIX || path.startsWith(`${API_PREFIX}/`);
  const url = prefixed ? path : `${API_PREFIX}${path}`;
  if (!params) {
    return url;
  }

  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined) {
      search.set(key, String(value));
    }
  }

  const query = search.toString();

  return query ? `${url}?${query}` : url;
}

async function parseErrorBody(response: Response): Promise<APIRequestError> {
  let body: APIErrorBody = {};
  try {
    body = (await response.json()) as APIErrorBody;
  } catch {
    // A failure outside the handler chain (a proxy, a truncated write) does not
    // carry the envelope. Fall through to the status text.
  }

  return new APIRequestError(
    response.status,
    body.error || response.statusText || "Request failed",
    body.fields ?? {},
  );
}

async function performRequest<T>(
  method: string,
  path: string,
  body?: unknown,
  options: RequestOptions = {},
): Promise<T> {
  const headers: Record<string, string> = { Accept: "application/json" };

  // nosurf only guards unsafe methods, and GET /api/* is required to stay
  // side-effect free, so the token rides on everything else.
  if (method !== "GET") {
    headers["X-CSRF-Token"] = csrfToken();
  }
  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
  }

  const response = await fetch(buildURL(path, options.params), {
    method,
    headers,
    credentials: "same-origin",
    signal: options.signal,
    body: body === undefined ? undefined : JSON.stringify(body),
  });

  // The API answers an unauthenticated request with 401 and no Location
  // precisely so this branch can exist: a redirect would be followed silently
  // and reach us as the login page's HTML under status 200.
  if (response.status === 401) {
    if (!options.skipAuthRedirect) {
      window.location.assign(LOGIN_PATH);
    }

    throw new APIRequestError(401, "Not signed in", {});
  }

  if (!response.ok) {
    throw await parseErrorBody(response);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
}

// Every call goes through here, so this is the one place the loading backdrop
// has to be driven from (lib/pending.ts) — a new route or a new resource gets
// the feedback without wiring anything, which is the failure the router-owned
// flag had. The counter spans the body read as well as the fetch, since a
// streamed response is still a request in flight.
//
// try/finally rather than a release after the await: a rejected fetch, an abort
// through `signal`, and the APIRequestError thrown for a non-2xx all have to
// release the counter, and one missed release leaves the backdrop covering the
// page for good.
async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  options: RequestOptions = {},
): Promise<T> {
  begin();
  try {
    return await performRequest<T>(method, path, body, options);
  } finally {
    end();
  }
}

export function get<T>(path: string, options?: RequestOptions): Promise<T> {
  return request<T>("GET", path, undefined, options);
}

export function post<T>(
  path: string,
  body?: unknown,
  options?: RequestOptions,
): Promise<T> {
  return request<T>("POST", path, body, options);
}

export function put<T>(
  path: string,
  body?: unknown,
  options?: RequestOptions,
): Promise<T> {
  return request<T>("PUT", path, body, options);
}

// `delete` is reserved, so the helper is del — the HTTP method it sends is not.
export function del<T>(path: string, options?: RequestOptions): Promise<T> {
  return request<T>("DELETE", path, undefined, options);
}
