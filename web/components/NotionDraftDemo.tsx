'use client';

import type React from 'react';
import { useCallback, useEffect, useMemo, useState } from 'react';

type NotionStatus = {
  configured: boolean;
  dry_run: boolean;
  database_id?: string;
};

type TaskRef = {
  id?: string;
  url?: string;
  dry_run: boolean;
};

type ErrorResponse = {
  error: string;
};

function formatError(e: unknown): string {
  if (typeof e === 'string') return e;
  if (e && typeof e === 'object' && 'message' in e) return String((e as any).message);
  return 'Unknown error';
}

export default function NotionDraftDemo() {
  const [title, setTitle] = useState('TEAmate demo draft');
  const [description, setDescription] = useState('This page was created from a TEAmate ticket draft.');

  const [status, setStatus] = useState<NotionStatus | null>(null);
  const [result, setResult] = useState<TaskRef | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const hasTitle = useMemo(() => title.trim().length > 0, [title]);

  const loadStatus = useCallback(async () => {
    setError(null);
    const res = await fetch('/api/notion/status', { cache: 'no-store' });
    const text = await res.text();
    if (!res.ok) {
      try {
        const parsed = JSON.parse(text) as ErrorResponse;
        throw new Error(parsed.error || `Request failed (${res.status})`);
      } catch {
        throw new Error(text || `Request failed (${res.status})`);
      }
    }
    setStatus(JSON.parse(text) as NotionStatus);
  }, []);

  const create = useCallback(async () => {
    if (!hasTitle) return;

    setLoading(true);
    setError(null);
    setResult(null);

    try {
      // Minimal TicketDraft payload (server keeps it as a pure transform; no DB needed).
      const payload = {
        title,
        description,
        labels: ['teammate', 'demo'],
        source_action_item_id: `ui-demo-${Date.now()}`,
        source_segment_ids: [],
      };

      const res = await fetch('/api/notion/pages', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      const text = await res.text();
      if (!res.ok) {
        try {
          const parsed = JSON.parse(text) as ErrorResponse;
          throw new Error(parsed.error || `Request failed (${res.status})`);
        } catch {
          throw new Error(text || `Request failed (${res.status})`);
        }
      }

      const parsed = JSON.parse(text) as TaskRef;
      setResult(parsed);
      await loadStatus();
    } catch (e) {
      setError(formatError(e));
    } finally {
      setLoading(false);
    }
  }, [description, hasTitle, loadStatus, title]);

  useEffect(() => {
    loadStatus().catch((e: unknown) => setError(formatError(e)));
  }, [loadStatus]);

  return (
    <section className="card" aria-label="notion draft demo">
      <h2 style={{ marginTop: 0 }}>Notion integration (ticket drafts → pages)</h2>

      <p style={{ marginTop: 0 }}>
        This calls <code>/api/notion/pages</code>, which proxies to the Go API at <code>/api/integrations/notion/pages</code>.
      </p>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        <div className="row" style={{ alignItems: 'center' }}>
          <strong>Status</strong>
          <small style={{ marginLeft: 12 }}>
            {status ? (
              <>
                configured: <code>{String(status.configured)}</code> · dry_run: <code>{String(status.dry_run)}</code>
              </>
            ) : (
              'loading…'
            )}
          </small>
          <button
            onClick={() => loadStatus().catch((e: unknown) => setError(formatError(e)))}
            disabled={loading}
            style={{ marginLeft: 'auto', background: 'rgba(255, 255, 255, 0.06)' }}
          >
            Refresh status
          </button>
        </div>

        <label>
          <div style={{ marginBottom: 6 }}>
            <strong>Draft title</strong>
          </div>
          <input value={title} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setTitle(e.target.value)} />
        </label>

        <label>
          <div style={{ marginBottom: 6 }}>
            <strong>Draft description</strong>
          </div>
          <textarea
            value={description}
            onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setDescription(e.target.value)}
          />
        </label>

        <div className="row">
          <button onClick={create} disabled={!hasTitle || loading}>
            {loading ? 'Creating…' : 'Create in Notion'}
          </button>
          <small style={{ marginLeft: 'auto' }}>
            Tip: set <code>NOTION_DRY_RUN=false</code> (and configure API/database) to create for real.
          </small>
        </div>

        {error ? (
          <div className="card" style={{ borderColor: 'rgba(255, 120, 120, 0.4)' }}>
            <strong>Request error</strong>
            <div style={{ marginTop: 8 }}>
              <small>{error}</small>
            </div>
          </div>
        ) : null}

        {result ? (
          <div className="card" style={{ borderColor: 'rgba(120, 255, 180, 0.25)' }}>
            <strong>Result</strong>
            <div style={{ marginTop: 8, display: 'flex', flexDirection: 'column', gap: 6 }}>
              <small>
                dry_run: <code>{String(result.dry_run)}</code>
              </small>
              {result.url ? (
                <small>
                  url: <a href={result.url} target="_blank" rel="noreferrer">{result.url}</a>
                </small>
              ) : (
                <small>url: (none)</small>
              )}
              {result.id ? (
                <small>
                  id: <code>{result.id}</code>
                </small>
              ) : null}
            </div>
          </div>
        ) : null}
      </div>
    </section>
  );
}
