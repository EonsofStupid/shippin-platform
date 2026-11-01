import assert from "node:assert/strict";
import http from "node:http";
import test from "node:test";
import { proxmoxService } from "./proxmox.svc.ts";

test("Proxmox client validates actual responses and action requests", async (t) => {
  let response = {
    configured: true,
    host: "example.test",
    port: 8006,
    node: "node1",
    reachable: true,
    message: "ready",
  };
  let code = 200;
  let raw = false;
  const requests = [];
  const server = http.createServer((req, res) => {
    requests.push({ method: req.method, url: req.url });
    res.writeHead(code, { "Content-Type": "application/json" });
    res.end(raw ? response : JSON.stringify(response));
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

  assert.equal((await proxmoxService.status()).reachable, true);
  response = { ...response, reachable: "true" };
  await assert.rejects(proxmoxService.status(), /Invalid Proxmox API response/);
  response = { count: 0, items: null };
  assert.deepEqual((await proxmoxService.guests()).items, []);
  response = { count: 1, items: null };
  await assert.rejects(proxmoxService.guests(), /Invalid Proxmox API response/);
  response = {
    count: 1,
    items: [
      { node: "node1", kind: "qemu", vmid: 101, name: "vm", status: "running", stable_label: "id" },
    ],
  };
  assert.equal((await proxmoxService.guests()).items[0].vmid, 101);
  response = {
    action_request_id: "id",
    kind: "guest_start",
    node: "node1",
    guest_kind: "qemu",
    vmid: 101,
    status: "succeeded",
  };
  assert.equal((await proxmoxService.start("node1", "qemu", 101)).status, "succeeded");
  assert.deepEqual(requests.at(-1), {
    method: "POST",
    url: "/api/warden/proxmox/guests/node1/qemu/101/start",
  });
  const before = requests.length;
  assert.throws(
    () => proxmoxService.stop("../node", "qemu", 101),
    /Invalid Proxmox guest reference/,
  );
  assert.throws(
    () => proxmoxService.stop("node1", "other", 101),
    /Invalid Proxmox guest reference/,
  );
  assert.equal(requests.length, before);
  code = 403;
  response = { error: { status: 403, message: "Access denied" } };
  await assert.rejects(proxmoxService.status(), /Access denied/);
  code = 200;
  raw = true;
  response = "not-json";
  await assert.rejects(proxmoxService.status(), /Invalid API response/);
});
