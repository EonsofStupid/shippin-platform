import { useEffect, useState } from "react";
import { api, type ClyffyHome } from "../../../../lib/api";
import { Grid, PageHeader, Stack } from "../../../../lib/design";
import { useAuth } from "../../identity";
import { MetricCard } from "../components/MetricCard";
import "./OverviewView.css";

/** OverviewView — the post-login operator landing: a compact, optimized
 *  dashboard of the foundation at a glance, on embossed coldlight surfaces. */
export function OverviewView() {
  const { token } = useAuth();
  const [workspaces, setWorkspaces] = useState<number | null>(null);
  const [home, setHome] = useState<ClyffyHome | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (token) {
      api
        .workspaces(token)
        .then((r) => setWorkspaces(r.count))
        .catch((e) => setError(e instanceof Error ? e.message : String(e)));
    } else {
      setWorkspaces(null);
    }
    api
      .clyffyHome()
      .then(setHome)
      .catch(() => undefined);
  }, [token]);

  const owed = home?.summary.config_items_owed ?? null;

  return (
    <Stack gap={5}>
      <PageHeader title="Overview" subtitle="WardenClyffe operator dashboard" />
      {error && (
        <p className="ov-error" role="alert">
          {error}
        </p>
      )}
      <Grid gap={4} min={14}>
        <MetricCard label="Workspaces" value={workspaces ?? "…"} tag="fleet" />
        <MetricCard label="Services" value={home?.summary.services ?? "…"} tag="mesh" tone="info" />
        <MetricCard label="Platforms" value={home?.summary.platforms ?? "…"} tag="infra" />
        <MetricCard
          label="Config owed"
          value={owed ?? "…"}
          tag="setup"
          tone={owed && owed > 0 ? "warning" : "success"}
        />
      </Grid>
    </Stack>
  );
}
