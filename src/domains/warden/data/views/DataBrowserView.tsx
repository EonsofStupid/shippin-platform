import { useEffect, useState } from "react";
import {
  api,
  type ColumnInfo,
  cell,
  type QueryResult,
  type RowPage,
  type SchemaInfo,
  type TableInfo,
} from "../../../../lib/api";
import { Badge, Button, Card, DataTable } from "../../../../lib/design";
import { useAuth } from "../../identity";

// Our own Go-backed, Supabase-inspired data browser.
export function DataBrowserView() {
  const { token } = useAuth();
  const [schemas, setSchemas] = useState<SchemaInfo[]>([]);
  const [schema, setSchema] = useState<string>("");
  const [tables, setTables] = useState<TableInfo[]>([]);
  const [table, setTable] = useState<string>("");
  const [cols, setCols] = useState<ColumnInfo[]>([]);
  const [page, setPage] = useState<RowPage | null>(null);
  const [sql, setSql] = useState(
    "select kind, state, count(*) from shippin_infra.resources group by 1,2;",
  );
  const [qres, setQres] = useState<QueryResult | null>(null);
  const [qerr, setQerr] = useState<string | null>(null);

  useEffect(() => {
    if (token)
      api
        .schemas(token)
        .then((r) => {
          setSchemas(r.items);
          if (r.items[0]) setSchema(r.items[0].name);
        })
        .catch((e) => setQerr((e as Error).message));
  }, [token]);
  useEffect(() => {
    if (schema && token)
      api
        .tables(token, schema)
        .then((r) => {
          setTables(r.items);
          setTable("");
          setPage(null);
          setCols([]);
        })
        .catch((e) => setQerr((e as Error).message));
  }, [schema, token]);

  async function openTable(t: string) {
    setTable(t);
    try {
      const [c, p] = await Promise.all([
        api.columns(token, schema, t),
        api.rows(token, schema, t, 50, 0),
      ]);
      setCols(c.columns);
      setPage(p);
      setQerr(null);
    } catch (e) {
      setQerr((e as Error).message);
    }
  }

  async function runQuery() {
    try {
      setQerr(null);
      setQres(await api.query(token, sql));
    } catch (e) {
      setQerr((e as Error).message);
      setQres(null);
    }
  }

  const pkSet = new Set(cols.filter((c) => c.is_primary_key).map((c) => c.name));

  return (
    <div>
      <h1 className="page-title">Data</h1>
      <p className="page-sub">
        Self-hosted Postgres, managed by our own Go layer — schemas, tables, rows, and read-only
        SQL.
      </p>

      <div
        style={{
          display: "grid",
          gridTemplateColumns: "clamp(11rem,16vw,15rem) 1fr",
          gap: "1rem",
          alignItems: "start",
        }}
      >
        {/* schema + table tree */}
        <Card title="Schemas">
          <div style={{ display: "flex", flexDirection: "column", gap: "0.25rem" }}>
            {schemas.map((s) => (
              <button
                type="button"
                key={s.name}
                className="navitem"
                data-active={s.name === schema}
                onClick={() => setSchema(s.name)}
              >
                {s.name} <span className="muted">· {s.tables}</span>
              </button>
            ))}
          </div>
        </Card>

        <div style={{ display: "grid", gap: "1rem" }}>
          <Card title={`Tables in ${schema || "—"}`}>
            <div className="row">
              {tables.map((t) => (
                <button
                  type="button"
                  key={t.name}
                  className="navitem"
                  data-active={t.name === table}
                  onClick={() => openTable(t.name)}
                >
                  {t.name}{" "}
                  <span className="muted">
                    · {t.columns} cols · ~{t.rows_estimate} rows
                  </span>
                </button>
              ))}
              {tables.length === 0 && <span className="muted">No tables.</span>}
            </div>
          </Card>

          {table && (
            <Card
              title={`${schema}.${table}`}
              action={<span className="muted">{page ? `${page.total} rows` : ""}</span>}
            >
              <div className="row" style={{ marginBottom: "0.75rem", gap: "0.5rem" }}>
                {cols.map((c) => (
                  <Badge key={c.name} dot={false} tone={c.is_primary_key ? "success" : "neutral"}>
                    {c.name}
                  </Badge>
                ))}
              </div>
              {page && (
                <DataTable
                  columns={page.columns}
                  rows={page.rows}
                  render={(v, col) => {
                    if (v === null || v === undefined) return <span className="ui-null">null</span>;
                    const s = cell(v);
                    return pkSet.has(col) ? <span className="ui-pk">{s}</span> : s;
                  }}
                />
              )}
            </Card>
          )}

          <Card title="SQL runner (read-only)" action={<Button onClick={runQuery}>Run</Button>}>
            <textarea
              value={sql}
              onChange={(e) => setSql(e.target.value)}
              spellCheck={false}
              className="mono"
              style={{
                width: "100%",
                minHeight: "5rem",
                background: "var(--wc-bg-elev)",
                color: "var(--wc-text)",
                border: "1px solid var(--wc-border)",
                borderRadius: "var(--wc-radius-sm)",
                padding: "0.5rem",
                fontSize: "var(--wc-fs-xs)",
              }}
            />
            {qerr && (
              <p className="muted" style={{ color: "var(--wc-danger)" }}>
                {qerr}
              </p>
            )}
            {qres && (
              <div style={{ marginTop: "0.75rem" }}>
                <DataTable
                  columns={qres.columns}
                  rows={qres.rows}
                  render={(v) => (v === null ? <span className="ui-null">null</span> : cell(v))}
                />
              </div>
            )}
          </Card>
        </div>
      </div>
    </div>
  );
}
