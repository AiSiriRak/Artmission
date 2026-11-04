import type { Page } from "@playwright/test";

// Records the requests that change data (POST, PUT, PATCH, DELETE) the page sends from now
// on, optionally only those whose path ends with `pathEnd`. A test uses it to prove that a
// rejected form sent nothing to the server.
export function recordWrites(page: Page, pathEnd = ""): string[] {
  const writes: string[] = [];
  page.on("request", (request) => {
    const isWrite = ["POST", "PUT", "PATCH", "DELETE"].includes(request.method());
    if (isWrite && new URL(request.url()).pathname.endsWith(pathEnd)) {
      writes.push(`${request.method()} ${request.url()}`);
    }
  });
  return writes;
}
