export async function api<T = Record<string, unknown>>(
  path: string,
  body?: unknown,
  signal?: AbortSignal,
  method?: "PUT",
): Promise<T> {
  const response = await fetch(path, {
    method: method ?? (body === undefined ? "GET" : "POST"),
    headers: {
      Accept: "application/json",
      ...(body === undefined ? {} : { "Content-Type": "application/json" }),
    },
    credentials: "same-origin",
    body: body === undefined ? undefined : JSON.stringify(body),
    signal,
  });
  if (response.status === 401) {
    window.location.assign("/login");
    throw new Error("Session expired");
  }
  const result = await response.json();
  if (!response.ok)
    throw new Error(
      result.detail ?? result.message ?? result.title ?? `HTTP error ${response.status}`,
    );
  return result as T;
}
