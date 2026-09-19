import { getPublicEnv } from "./env";

/**
 * Thrown for any non-2xx response from the Nexus API. Always carries the
 * API's own stable error code (never a raw HTTP status alone), so callers
 * can branch on e.g. "ASSET_NOT_FOUND" without parsing prose.
 */
export class ApiError extends Error {
  readonly code: string;
  readonly status: number;

  constructor(code: string, message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
  }
}

interface ApiErrorBody {
  readonly error: { readonly code: string; readonly message: string };
}

/**
 * Calls the Go API and decodes its JSON body. Never swallows a non-2xx
 * response as if it were data — every failure becomes a typed {@link ApiError}
 * carrying the API's own error code and message, never a fabricated one.
 */
export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const { apiBaseUrl } = getPublicEnv();
  const response = await fetch(`${apiBaseUrl}${path}`, {
    ...init,
    headers: { Accept: "application/json", ...init?.headers },
  });

  if (!response.ok) {
    let body: ApiErrorBody | undefined;
    try {
      body = (await response.json()) as ApiErrorBody;
    } catch {
      // Response body wasn't JSON (e.g. a proxy error page) — fall through
      // to the generic error below rather than fabricating a code.
    }
    if (body?.error) {
      throw new ApiError(body.error.code, body.error.message, response.status);
    }
    throw new ApiError("UNKNOWN_ERROR", `Request to ${path} failed with HTTP ${response.status}.`, response.status);
  }

  return (await response.json()) as T;
}

/** Builds a query string from the given params, omitting undefined/empty values. */
export function toQueryString(params: Record<string, string | number | boolean | undefined>): string {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== "") {
      search.set(key, String(value));
    }
  }
  const query = search.toString();
  return query ? `?${query}` : "";
}
