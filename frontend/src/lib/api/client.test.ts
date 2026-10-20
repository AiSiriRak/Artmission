const mockGetAccessToken = jest.fn<string | null, []>();
const mockSetAccessToken = jest.fn<void, [string]>();
const mockClearAccessToken = jest.fn<void, []>();

jest.mock("@/lib/token", () => ({
  getAccessToken: mockGetAccessToken,
  setAccessToken: mockSetAccessToken,
  clearAccessToken: mockClearAccessToken,
}));

function jsonResponse(status: number, body?: unknown): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    text: async () => (body === undefined ? "" : JSON.stringify(body)),
  } as Response;
}

// client.ts keeps a module-level `refreshPromise` singleton for the
// single-flight dedupe. Each test needs a clean instance of that state, so
// the module is dynamically re-imported after `jest.resetModules()` per
// test instead of statically imported once.
describe("apiFetch 401 refresh interceptor", () => {
  const originalFetch = global.fetch;
  const originalApiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL;
  let fetchMock: jest.Mock;

  beforeEach(() => {
    jest.resetModules();
    process.env.NEXT_PUBLIC_API_BASE_URL = "http://api.test";
    mockGetAccessToken.mockReset().mockReturnValue("stale-token");
    mockSetAccessToken.mockReset();
    mockClearAccessToken.mockReset();
    fetchMock = jest.fn();
    global.fetch = fetchMock as unknown as typeof fetch;
  });

  afterEach(() => {
    global.fetch = originalFetch;
    process.env.NEXT_PUBLIC_API_BASE_URL = originalApiBaseUrl;
  });

  it("returns the response directly when the request succeeds", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse(200, { ok: true }));

    const { apiFetch } = await import("./client");
    const result = await apiFetch<{ ok: boolean }>("/widgets");

    expect(result).toEqual({ ok: true });
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("refreshes once on 401 and retries the original request", async () => {
    fetchMock
      .mockResolvedValueOnce(jsonResponse(401, { message: "expired" }))
      .mockResolvedValueOnce(jsonResponse(200, { access_token: "fresh-token" }))
      .mockResolvedValueOnce(jsonResponse(200, { ok: true }));

    const { apiFetch } = await import("./client");
    const result = await apiFetch<{ ok: boolean }>("/widgets");

    expect(result).toEqual({ ok: true });
    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(fetchMock.mock.calls[1][0]).toBe("http://api.test/auth/refresh");
    expect(mockSetAccessToken).toHaveBeenCalledWith("fresh-token");
  });

  it("dedupes concurrent 401s into a single refresh call", async () => {
    fetchMock.mockImplementation((url: string) => {
      if (url.endsWith("/auth/refresh")) {
        return Promise.resolve(
          jsonResponse(200, { access_token: "fresh-token" }),
        );
      }
      if (url.endsWith("/a")) {
        return fetchMock.mock.calls.filter((c) => c[0] === url).length === 1
          ? Promise.resolve(jsonResponse(401))
          : Promise.resolve(jsonResponse(200, { from: "a" }));
      }
      return fetchMock.mock.calls.filter((c) => c[0] === url).length === 1
        ? Promise.resolve(jsonResponse(401))
        : Promise.resolve(jsonResponse(200, { from: "b" }));
    });

    const { apiFetch } = await import("./client");
    const [a, b] = await Promise.all([
      apiFetch<{ from: string }>("/a"),
      apiFetch<{ from: string }>("/b"),
    ]);

    expect(a).toEqual({ from: "a" });
    expect(b).toEqual({ from: "b" });

    const refreshCalls = fetchMock.mock.calls.filter(
      (c) => c[0] === "http://api.test/auth/refresh",
    );
    expect(refreshCalls).toHaveLength(1);
  });

  it("clears the token and rejects with 401 when refresh itself fails", async () => {
    fetchMock
      .mockResolvedValueOnce(jsonResponse(401, { message: "expired" }))
      .mockResolvedValueOnce(jsonResponse(401, { message: "invalid refresh" }))
      .mockResolvedValueOnce(jsonResponse(401, { message: "expired" }))
      .mockResolvedValueOnce(jsonResponse(401, { message: "invalid refresh" }));

    const { apiFetch, ApiError } = await import("./client");

    await expect(apiFetch("/widgets")).rejects.toMatchObject({
      status: 401,
    });
    await expect(apiFetch("/widgets")).rejects.toBeInstanceOf(ApiError);
    expect(fetchMock).toHaveBeenCalledTimes(4); // 2 original attempts + 2 refresh attempts
    expect(mockClearAccessToken).toHaveBeenCalled();
    expect(mockSetAccessToken).not.toHaveBeenCalled();
  });

  it("does not attempt a refresh for auth endpoints themselves", async () => {
    fetchMock.mockResolvedValueOnce(
      jsonResponse(401, { message: "bad credentials" }),
    );

    const { apiFetch } = await import("./client");

    await expect(
      apiFetch("/auth/login", { method: "POST" }),
    ).rejects.toMatchObject({ status: 401 });
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
});
