# Devstation turnkey package — mirror of the installed authority

These files are the **version-controlled mirror of the secret-access pattern
already installed on `warden-devstation-01`**, so a new customer devstation can
be provisioned identically. They are not a separate design — the **authority is
`W:/configuration`** (`configuration/providers/infisical.yaml`,
`runtime-keyring.yaml`, `authority.yaml`). Do not fork a second pattern.

```text
turnkey/
  bin/
    warden-infisical-bootstrap   /usr/local/bin/  — encrypts the MI creds with
                                 systemd-creds into /etc/credstore.encrypted,
                                 starts the agent, waits for the token sink
    warden-infisical-status      /usr/local/bin/  — the Infisical state machine
                                 (…credentials_missing → …active)
    warden-devstation-status     /usr/local/bin/  — host/workspace/tools/git
  etc/warden/infisical/
    agent-config.yaml            /etc/warden/infisical/  — universal-auth reading
                                 the systemd-decrypted creds at /run/credentials,
                                 token sink /run/shippin-secrets/infisical-access-token
  systemd/
    warden-infisical-agent.service  LoadCredentialEncrypted; runs as wardenop
```

## How auth actually works (one boring, predictable path)

1. Secrets are **never** on W and **never** in plaintext on disk. The machine
   identity client id/secret are encrypted at rest with `systemd-creds
   --with-key=host` under `/etc/credstore.encrypted/` (`0600`, root).
2. `warden-infisical-agent.service` decrypts them into
   `/run/credentials/warden-infisical-agent.service/` and the Infisical agent
   exchanges them for a token at `/run/shippin-secrets/infisical-access-token`.
3. Process secrets live only on the volatile runtime keyring
   `/run/shippin-secrets` (`0700`, reboot-cleared), managed by the
   `shippin-secret-*` helpers (`warden-secret-*` is a compatibility wrap).
   `static_pat` is **break-glass only**. `/run/warden-secrets` is a symlink.

## The one operator action

Status today is `machine_identity_credentials_missing`. To activate:

```bash
sudo warden-infisical-bootstrap     # prompts for a LIVE client id + secret
warden-infisical-status             # → status=machine_identity_active
```

or drive it from the Warden UI: **/admin/connect** (Connect & Launch) feeds the
same `warden-infisical-bootstrap` non-interactively. Nothing else is missing —
the whole pipeline is installed; it just needs a live credential.
