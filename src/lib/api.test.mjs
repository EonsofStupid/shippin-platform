import assert from "node:assert/strict";
import http from "node:http";
import test from "node:test";
import { api } from "./api.ts";

test("staff data and fleet requests carry the operator session", async (t) => {
  const requests = [];
  const server = http.createServer((request, response) => {
    requests.push({ url: request.url, authorization: request.headers.authorization });
    response.writeHead(200, { "content-type": "application/json" });
    response.end("{}");
  });
  await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
  const originalFetch = globalThis.fetch;
  globalThis.fetch = (path, options) =>
    originalFetch(`http://127.0.0.1:${server.address().port}${path}`, options);
  t.after(async () => {
    globalThis.fetch = originalFetch;
    await new Promise((resolve, reject) =>
      server.close((error) => (error ? reject(error) : resolve())),
    );
  });

  await api.workspaces("session");
  await api.provision("session", { tier: "test" });
  await api.schemas("session");
  await api.tables("session", "shippin_core");
  await api.columns("session", "shippin_core", "tenants");
  await api.rows("session", "shippin_core", "tenants");
  await api.query("session", "select 1");
  assert.equal(requests.length, 7);
  assert.ok(requests.every((request) => request.authorization === "Bearer session"));
  await assert.rejects(api.workspaces(null), /Operator authentication is required/);
  assert.equal(requests.length, 7);
});
