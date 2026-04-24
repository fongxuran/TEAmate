'use client';

import type React from 'react';
import { useCallback, useEffect, useMemo, useState } from 'react';

type Message = {
  id: number;
  body: string;
  created_at: string;
};

type ListMessagesResponse = {
  messages: Message[];
  limit: number;
  offset: number;
};

type ErrorResponse = {
  error: string;
};

type ErrorWithMessage = {
  message: string;
};

function isErrorWithMessage(value: unknown): value is ErrorWithMessage {
  return (
    typeof value === 'object' &&
    value !== null &&
    'message' in value &&
    typeof (value as { message?: unknown }).message === 'string'
  );
}

function formatError(e: unknown): string {
  if (typeof e === 'string') return e;
  if (isErrorWithMessage(e)) return e.message;
  return 'Unknown error';
}

export default function MessageDemo() {
  const [body, setBody] = useState('');
  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const hasBody = useMemo(() => body.trim().length > 0, [body]);

  const loadMessages = useCallback(async () => {
    setError(null);
    const res = await fetch('/api/messages?limit=20&offset=0', { cache: 'no-store' });
    const text = await res.text();

    if (!res.ok) {
      try {
        const parsed = JSON.parse(text) as ErrorResponse;
        throw new Error(parsed.error || `Request failed (${res.status})`);
      } catch {
        throw new Error(text || `Request failed (${res.status})`);
      }
    }

    const parsed = JSON.parse(text) as ListMessagesResponse;
    setMessages(parsed.messages);
  }, []);

  const submit = useCallback(async () => {
    if (!hasBody) return;

    setLoading(true);
    setError(null);

    try {
      const res = await fetch('/api/messages', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ body }),
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

      setBody('');
      await loadMessages();
    } catch (e) {
      setError(formatError(e));
    } finally {
      setLoading(false);
    }
  }, [body, hasBody, loadMessages]);

  useEffect(() => {
    loadMessages().catch((e: unknown) => setError(formatError(e)));
  }, [loadMessages]);

  return (
    <section className="card" aria-label="message demo">
      <h2 style={{ marginTop: 0 }}>Backend connectivity (messages)</h2>
      <p style={{ marginTop: 0 }}>
        This uses a Next.js route handler at <code>/api/messages</code> to proxy requests to the Go API.
      </p>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        <label>
          <div style={{ marginBottom: 6 }}>
            <strong>New message</strong>
          </div>
          <textarea
            value={body}
            onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setBody(e.target.value)}
            placeholder="Say hello to the backend…"
          />
        </label>

        <div className="row">
          <button onClick={submit} disabled={!hasBody || loading}>
            {loading ? 'Sending…' : 'Send'}
          </button>
          <button
            onClick={() => loadMessages().catch((e: unknown) => setError(formatError(e)))}
            disabled={loading}
            style={{ background: 'rgba(255, 255, 255, 0.06)' }}
          >
            Refresh
          </button>
          <small style={{ marginLeft: 'auto' }}>{messages.length} messages</small>
        </div>

        {error ? (
          <div className="card" style={{ borderColor: 'rgba(255, 120, 120, 0.4)' }}>
            <strong>Request error</strong>
            <div style={{ marginTop: 8 }}>
              <small>{error}</small>
            </div>
          </div>
        ) : null}

        <hr />

        <ol style={{ margin: 0, paddingLeft: 18, display: 'flex', flexDirection: 'column', gap: 8 }}>
          {messages.map((m) => (
            <li key={m.id}>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
                <div>{m.body}</div>
                <small>
                  #{m.id} · {new Date(m.created_at).toLocaleString()}
                </small>
              </div>
            </li>
          ))}
        </ol>
      </div>
    </section>
  );
}
