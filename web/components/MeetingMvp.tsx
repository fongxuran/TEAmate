'use client';

import type React from 'react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

type AgendaItem = {
  id: string;
  title: string;
  description?: string | null;
  keywords?: string[];
};

type Transcript = {
  meeting_id?: string | null;
  meeting_name?: string | null;
  turns: Array<{ timestamp?: string | null; speaker?: string | null; text: string }>;
};

type Segment = {
  segment_id: string;
  start_turn_idx: number;
  end_turn_idx: number;
  text: string;
  speaker_distribution?: Record<string, number>;
};

type DriftSegment = {
  segment: Segment;
  best_agenda_item_id?: string | null;
  best_agenda_title?: string | null;
  best_score: number;
  is_drift: boolean;
  feedback_override?: boolean | null;
};

type ActionItem = {
  action_item_id: string;
  title: string;
  description?: string | null;
  owner?: string | null;
  due_date?: string | null;
  source_segment_ids?: string[];
  confidence: number;
};

type TicketDraft = {
  title: string;
  description: string;
  labels?: string[];
  owner?: string | null;
  priority?: string | null;
  source_action_item_id: string;
  source_meeting_id?: string | null;
  source_meeting_name?: string | null;
  source_segment_ids?: string[];
};

type AnalysisResult = {
  schema_version: string;
  generated_at: string;
  agenda: AgendaItem[];
  transcript: Transcript;
  segments: DriftSegment[];
  summary: string;
  decisions: string[];
  action_items: ActionItem[];
  ticket_drafts: TicketDraft[];
};

type Config = {
  provider: string;
  drift_threshold: number;
  segment_max_tokens: number;
  segment_max_chars: number;
};

type RealtimeMessage = {
  client_id: string;
  timestamp: string;
  text_delta?: string | null;
  text?: string | null;
  author?: string | null;
};

type DriftAlert = {
  segment_id: string;
  best_agenda_title?: string | null;
  best_score: number;
  text_preview: string;
};

type SyncPayload = {
  session_id: string;
  agenda_text: string;
  transcript_text: string;
  agenda: AgendaItem[];
  config: Config;
  analysis: AnalysisResult;
};

type WsEvent =
  | { type: 'sync'; payload: SyncPayload }
  | { type: 'agenda_updated'; payload: SyncPayload }
  | { type: 'transcript_updated'; payload: RealtimeMessage }
  | { type: 'analysis_updated'; payload: AnalysisResult }
  | { type: 'drift_alert'; payload: DriftAlert }
  | { type: 'drift_feedback_applied'; payload: { segment_id: string; is_drift: boolean } }
  | { type: 'reset_applied'; payload: SyncPayload };

type ExportResponse = {
  schema_version: string;
  generated_at: string;
  drafts: TicketDraft[];
  markdown: string;
  csv: string;
};

type ErrorWithMessage = {
  message: string;
};

type UploadedAgendaItem = {
  title?: string | null;
};

type UploadedTranscriptTurn = {
  speaker?: string | null;
  text?: string | null;
};

type UploadedTranscript = {
  turns?: UploadedTranscriptTurn[];
};

type UploadedMeeting = {
  schema_version?: string;
  agenda?: UploadedAgendaItem[];
  transcript?: UploadedTranscript;
};

function isErrorWithMessage(value: unknown): value is ErrorWithMessage {
  return (
    typeof value === 'object' &&
    value !== null &&
    'message' in value &&
    typeof (value as { message?: unknown }).message === 'string'
  );
}

function isUploadedMeeting(value: unknown): value is UploadedMeeting {
  if (!value || typeof value !== 'object') return false;
  const record = value as Record<string, unknown>;
  if (record.schema_version !== 'v1') return false;
  const transcript = record.transcript;
  if (!transcript || typeof transcript !== 'object') return false;
  const turns = (transcript as Record<string, unknown>).turns;
  return Array.isArray(turns);
}

function formatError(e: unknown): string {
  if (typeof e === 'string') return e;
  if (isErrorWithMessage(e)) return e.message;
  return 'Unknown error';
}

function downloadText(filename: string, content: string, mime: string) {
  const blob = new Blob([content], { type: mime });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

function defaultWsUrl(): string {
  return process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws';
}

export default function MeetingMvp() {
  const [sessionId, setSessionId] = useState('default');
  const [clientId, setClientId] = useState<string>('');

  const [agendaText, setAgendaText] = useState('');
  const [transcriptText, setTranscriptText] = useState('');
  const [config, setConfig] = useState<Config>({ provider: 'deterministic', drift_threshold: 0.08, segment_max_tokens: 220, segment_max_chars: 1800 });

  const [analysis, setAnalysis] = useState<AnalysisResult | null>(null);
  const [driftAlert, setDriftAlert] = useState<DriftAlert | null>(null);
  const [showOnlyDrift, setShowOnlyDrift] = useState(false);

  const [wsState, setWsState] = useState<'disconnected' | 'connecting' | 'connected'>('disconnected');
  const wsRef = useRef<WebSocket | null>(null);
  const transcriptRef = useRef<string>('');

  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const agendaSendTimer = useRef<number | null>(null);
  const configSendTimer = useRef<number | null>(null);

  useEffect(() => {
    setClientId(globalThis.crypto?.randomUUID?.() ?? `client-${Math.random().toString(16).slice(2)}`);
  }, []);

  useEffect(() => {
    transcriptRef.current = transcriptText;
  }, [transcriptText]);

  const connected = wsState === 'connected';

  const send = useCallback((type: string, payload: unknown) => {
    const ws = wsRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) return;
    ws.send(JSON.stringify({ type, payload }));
  }, []);

  const connect = useCallback(() => {
    if (!clientId) return;
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) return;

    setError(null);
    setWsState('connecting');

    const base = defaultWsUrl();
    const url = new URL(base);
    url.searchParams.set('session', sessionId || 'default');
    url.searchParams.set('client_id', clientId);

    const ws = new WebSocket(url.toString());
    wsRef.current = ws;

    ws.onopen = () => {
      setWsState('connected');
    };

    ws.onclose = () => {
      setWsState('disconnected');
      wsRef.current = null;
    };

    ws.onerror = () => {
      setError('WebSocket error');
    };

    ws.onmessage = (ev: MessageEvent<string>) => {
      try {
        const parsed = JSON.parse(ev.data) as WsEvent;
        if (parsed.type === 'sync' || parsed.type === 'agenda_updated' || parsed.type === 'reset_applied') {
          setAgendaText(parsed.payload.agenda_text || '');
          setTranscriptText(parsed.payload.transcript_text || '');
          setConfig((prev) => ({ ...prev, ...(parsed.payload.config || {}) }));
          setAnalysis(parsed.payload.analysis || null);
          return;
        }

        if (parsed.type === 'transcript_updated') {
          const msg = parsed.payload;
          setTranscriptText((prev: string) => {
            if (msg.text_delta) {
              // Avoid double-applying optimistic local append.
              if (msg.client_id === clientId && prev.endsWith(msg.text_delta)) return prev;
              return prev + msg.text_delta;
            }
            if (typeof msg.text === 'string') return msg.text;
            return prev;
          });
          return;
        }

        if (parsed.type === 'analysis_updated') {
          setAnalysis(parsed.payload);
          return;
        }

        if (parsed.type === 'drift_alert') {
          setDriftAlert(parsed.payload);
          return;
        }

        if (parsed.type === 'drift_feedback_applied') {
          // analysis_updated will typically follow; keep this for a quick UX affordance.
          setDriftAlert(null);
        }
      } catch (e) {
        setError(`WS parse error: ${formatError(e)}`);
      }
    };
  }, [clientId, config, sessionId]);

  const disconnect = useCallback(() => {
    const ws = wsRef.current;
    if (!ws) return;
    ws.close();
  }, []);

  const setAgendaTextAndBroadcast = useCallback(
    (value: string) => {
      setAgendaText(value);
      if (agendaSendTimer.current) window.clearTimeout(agendaSendTimer.current);
      agendaSendTimer.current = window.setTimeout(() => {
        send('set_agenda', { agenda_text: value });
      }, 250);
    },
    [send]
  );

  const setConfigAndBroadcast = useCallback(
    (next: Config) => {
      setConfig(next);
      if (configSendTimer.current) window.clearTimeout(configSendTimer.current);
      configSendTimer.current = window.setTimeout(() => {
        send('set_config', next);
      }, 250);
    },
    [send]
  );

  const setTranscriptAndBroadcast = useCallback(
    (nextText: string) => {
      setTranscriptText(nextText);

      // Append-only (delta) when possible; fallback to replace.
      const prev = transcriptRef.current;
      if (nextText.startsWith(prev)) {
        const delta = nextText.slice(prev.length);
        if (delta.length > 0) {
          send('realtime_message', {
            message: { text_delta: delta, author: 'local' },
          });
        }
      } else {
        send('realtime_message', {
          message: { text: nextText, author: 'local' },
        });
      }
    },
    [send]
  );

  const analyzeNow = useCallback(async () => {
    setBusy(true);
    setError(null);
    try {
      const res = await fetch(`/api/meeting/analyze?session=${encodeURIComponent(sessionId || 'default')}`, {
        method: 'POST',
      });
      const text = await res.text();
      if (!res.ok) throw new Error(text || `Analyze failed (${res.status})`);
      setAnalysis(JSON.parse(text) as AnalysisResult);
    } catch (e) {
      setError(formatError(e));
    } finally {
      setBusy(false);
    }
  }, [sessionId]);

  const exportDrafts = useCallback(async () => {
    setBusy(true);
    setError(null);
    try {
      const res = await fetch(`/api/exports/ticket-drafts?session=${encodeURIComponent(sessionId || 'default')}`, {
        method: 'POST',
      });
      const text = await res.text();
      if (!res.ok) throw new Error(text || `Export failed (${res.status})`);
      const parsed = JSON.parse(text) as ExportResponse;

      downloadText('ticket_drafts.json', JSON.stringify({ schema_version: 'v1', generated_at: parsed.generated_at, drafts: parsed.drafts }, null, 2) + '\n', 'application/json');
      downloadText('ticket_drafts.md', parsed.markdown, 'text/markdown');
      downloadText('ticket_drafts.csv', parsed.csv, 'text/csv');
    } catch (e) {
      setError(formatError(e));
    } finally {
      setBusy(false);
    }
  }, [sessionId]);

  const resetSession = useCallback(() => {
    send('reset', {});
  }, [send]);

  const applyDriftFeedback = useCallback(
    (segmentId: string, isDrift: boolean) => {
      send('drift_feedback', { segment_id: segmentId, is_drift: isDrift });
      setDriftAlert(null);
    },
    [send]
  );

  const filteredSegments = useMemo(() => {
    const segs = analysis?.segments ?? [];
    return showOnlyDrift ? segs.filter((s: DriftSegment) => s.is_drift) : segs;
  }, [analysis, showOnlyDrift]);

  const loadSampleMeeting1 = useCallback(async () => {
    setBusy(true);
    setError(null);
    try {
      const res = await fetch('/api/samples/meeting-1', { cache: 'no-store' });
      const text = await res.text();
      if (!res.ok) throw new Error(text || `Load sample failed (${res.status})`);
      setTranscriptAndBroadcast(text);
    } catch (e) {
      setError(formatError(e));
    } finally {
      setBusy(false);
    }
  }, [setTranscriptAndBroadcast]);

  const onUpload = useCallback(
    async (file: File | null) => {
      if (!file) return;
      setBusy(true);
      setError(null);
      try {
        const raw = await file.text();
        if (file.name.toLowerCase().endsWith('.json')) {
          const parsed = JSON.parse(raw) as unknown;
          if (isUploadedMeeting(parsed)) {
            const agendaLines = Array.isArray(parsed.agenda)
              ? parsed.agenda
                  .map((item) => (typeof item?.title === 'string' ? item.title : ''))
                  .filter((title) => title.trim().length > 0)
              : [];
            if (agendaLines.length > 0) {
              setAgendaTextAndBroadcast(agendaLines.join('\n'));
            }
            const turns = parsed.transcript?.turns ?? [];
            const lines = turns
              .map((turn) => {
                const spk = typeof turn?.speaker === 'string' && turn.speaker.trim() ? `${turn.speaker.trim()}: ` : '';
                const txt = typeof turn?.text === 'string' ? turn.text.trim() : '';
                return spk + txt;
              })
              .filter((l) => l.trim().length > 0);
            setTranscriptAndBroadcast(lines.join('\n'));
            return;
          }
        }

        // Fallback: treat as plain text transcript.
        setTranscriptAndBroadcast(raw);
      } catch (e) {
        setError(formatError(e));
      } finally {
        setBusy(false);
      }
    },
    [setAgendaTextAndBroadcast, setTranscriptAndBroadcast]
  );

  return (
    <section className="card" aria-label="local mvp">
      <div className="row" style={{ justifyContent: 'space-between' }}>
        <h2 style={{ margin: 0 }}>Local MVP (T-011)</h2>
        <small>
          WS: <code>{defaultWsUrl()}</code>
        </small>
      </div>

      <p style={{ marginTop: 8 }}>
        Two-browser demo: open this page in two tabs/windows, click <strong>Connect</strong> in both, then type into the
        transcript box. Drift alerts + feedback should propagate.
      </p>

      {driftAlert ? (
        <div className="card" style={{ borderColor: 'rgba(124, 255, 214, 0.35)' }}>
          <div style={{ display: 'flex', gap: 12, alignItems: 'flex-start' }}>
            <div style={{ flex: 1 }}>
              <strong>Drift alert</strong>
              <div style={{ marginTop: 6 }}>
                <small>
                  Segment <code>{driftAlert.segment_id}</code> · best match: {driftAlert.best_agenda_title ?? '—'} · score:{' '}
                  {driftAlert.best_score.toFixed(3)}
                </small>
              </div>
              <div style={{ marginTop: 10, whiteSpace: 'pre-wrap' }}>{driftAlert.text_preview}</div>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              <button onClick={() => applyDriftFeedback(driftAlert.segment_id, true)} disabled={!connected}>
                Drift
              </button>
              <button
                onClick={() => applyDriftFeedback(driftAlert.segment_id, false)}
                disabled={!connected}
                style={{ background: 'rgba(255, 255, 255, 0.06)' }}
              >
                Not drift
              </button>
            </div>
          </div>
        </div>
      ) : null}

      {error ? (
        <div className="card" style={{ borderColor: 'rgba(255, 120, 120, 0.4)', marginTop: 12 }}>
          <strong>Error</strong>
          <div style={{ marginTop: 8 }}>
            <small>{error}</small>
          </div>
        </div>
      ) : null}

      <hr />

      <section aria-label="inputs" style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <h3 style={{ margin: 0 }}>Inputs</h3>

        <div className="row">
          <label style={{ flex: 1 }}>
            <small>Session</small>
            <input value={sessionId} onChange={(e) => setSessionId(e.target.value)} placeholder="default" />
          </label>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <small>Realtime</small>
            <div className="row">
              <button onClick={connect} disabled={wsState !== 'disconnected'}>
                {wsState === 'connecting' ? 'Connecting…' : 'Connect'}
              </button>
              <button
                onClick={disconnect}
                disabled={wsState !== 'connected'}
                style={{ background: 'rgba(255, 255, 255, 0.06)' }}
              >
                Disconnect
              </button>
            </div>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <small>Actions</small>
            <div className="row">
              <button onClick={analyzeNow} disabled={busy}>
                Analyze now
              </button>
              <button onClick={resetSession} disabled={!connected} style={{ background: 'rgba(255, 120, 120, 0.14)' }}>
                Reset session
              </button>
            </div>
          </div>
        </div>

        <label>
          <div style={{ marginBottom: 6 }}>
            <strong>Agenda</strong> <small>(shared)</small>
          </div>
          <textarea
            value={agendaText}
            onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setAgendaTextAndBroadcast(e.target.value)}
            placeholder="- Status\n- Decisions\n- Action items"
          />
        </label>

        <label>
          <div style={{ marginBottom: 6 }}>
            <strong>Transcript</strong> <small>(shared, realtime)</small>
          </div>
          <textarea
            value={transcriptText}
            onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setTranscriptAndBroadcast(e.target.value)}
            placeholder="Paste or type transcript here…"
          />
          <div className="row" style={{ marginTop: 8, justifyContent: 'space-between' }}>
            <small>
              Client: <code>{clientId || '…'}</code> · {connected ? 'connected' : 'disconnected'}
            </small>
            <div className="row">
              <button onClick={loadSampleMeeting1} disabled={busy} style={{ background: 'rgba(255, 255, 255, 0.06)' }}>
                Load sample (meeting 1)
              </button>
              <label style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                <small>Upload</small>
                <input
                  type="file"
                  accept=".txt,.json"
                  onChange={(e) => onUpload(e.target.files?.[0] ?? null)}
                  style={{ width: 220 }}
                />
              </label>
            </div>
          </div>
        </label>

        <div className="row">
          <label style={{ flex: 1 }}>
            <small>Provider</small>
            <select value={config.provider} onChange={(e) => setConfigAndBroadcast({ ...config, provider: e.target.value })}>
              <option value="deterministic">Deterministic (local)</option>
              <option value="anthropic">Anthropic (placeholder)</option>
            </select>
          </label>
          <label style={{ flex: 1 }}>
            <small>Drift threshold</small>
            <input
              type="number"
              step="0.01"
              value={config.drift_threshold}
              onChange={(e) => setConfigAndBroadcast({ ...config, drift_threshold: Number(e.target.value) })}
            />
          </label>
          <label style={{ flex: 1 }}>
            <small>Segment max tokens</small>
            <input
              type="number"
              value={config.segment_max_tokens}
              onChange={(e) => setConfigAndBroadcast({ ...config, segment_max_tokens: Number(e.target.value) })}
            />
          </label>
          <label style={{ flex: 1 }}>
            <small>Segment max chars</small>
            <input
              type="number"
              value={config.segment_max_chars}
              onChange={(e) => setConfigAndBroadcast({ ...config, segment_max_chars: Number(e.target.value) })}
            />
          </label>
        </div>
      </section>

      <hr />

      <section aria-label="drift" style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div className="row" style={{ justifyContent: 'space-between' }}>
          <h3 style={{ margin: 0 }}>Drift view</h3>
          <label className="row" style={{ gap: 8 }}>
            <input type="checkbox" checked={showOnlyDrift} onChange={(e) => setShowOnlyDrift(e.target.checked)} />
            <small>Show drift only</small>
          </label>
        </div>

        {analysis ? (
          <small>
            {analysis.segments.length} segments · {analysis.segments.filter((s) => s.is_drift).length} drift
          </small>
        ) : (
          <small>No analysis yet. Connect + type, or click “Analyze now”.</small>
        )}

        <ol style={{ margin: 0, paddingLeft: 18, display: 'flex', flexDirection: 'column', gap: 10 }}>
          {filteredSegments.map((s) => (
            <li key={s.segment.segment_id}>
              <div className="card" style={{ padding: 12 }}>
                <div className="row" style={{ justifyContent: 'space-between' }}>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
                    <strong>
                      <code>{s.segment.segment_id}</code> · {s.is_drift ? 'Drift' : 'On track'}
                      {s.feedback_override !== null && s.feedback_override !== undefined ? (
                        <small> (override)</small>
                      ) : null}
                    </strong>
                    <small>
                      best: {s.best_agenda_title ?? '—'} · score {s.best_score.toFixed(3)} · turns {s.segment.start_turn_idx}…
                      {s.segment.end_turn_idx}
                    </small>
                  </div>
                  <div className="row">
                    <button onClick={() => applyDriftFeedback(s.segment.segment_id, true)} disabled={!connected}>
                      Drift
                    </button>
                    <button
                      onClick={() => applyDriftFeedback(s.segment.segment_id, false)}
                      disabled={!connected}
                      style={{ background: 'rgba(255, 255, 255, 0.06)' }}
                    >
                      Not drift
                    </button>
                  </div>
                </div>
                <div style={{ marginTop: 10, whiteSpace: 'pre-wrap' }}>{s.segment.text}</div>
              </div>
            </li>
          ))}
        </ol>
      </section>

      <hr />

      <section aria-label="outcomes" style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <h3 style={{ margin: 0 }}>Outcomes</h3>

        <div className="card" style={{ padding: 12 }}>
          <strong>Summary</strong>
          <div style={{ marginTop: 8 }}>{analysis?.summary || <small>(none)</small>}</div>
        </div>

        <div className="row" style={{ gap: 12, alignItems: 'stretch' }}>
          <div className="card" style={{ flex: 1, padding: 12 }}>
            <strong>Decisions</strong>
            <ul style={{ margin: 8, paddingLeft: 18 }}>
              {(analysis?.decisions ?? []).map((d, i) => (
                <li key={i}>{d}</li>
              ))}
              {(analysis?.decisions ?? []).length === 0 ? <small>(none)</small> : null}
            </ul>
          </div>

          <div className="card" style={{ flex: 1, padding: 12 }}>
            <strong>Action items</strong>
            <ul style={{ margin: 8, paddingLeft: 18 }}>
              {(analysis?.action_items ?? []).map((a) => (
                <li key={a.action_item_id}>
                  {a.title} <small>({a.confidence.toFixed(2)})</small>
                </li>
              ))}
              {(analysis?.action_items ?? []).length === 0 ? <small>(none)</small> : null}
            </ul>
          </div>
        </div>
      </section>

      <hr />

      <section aria-label="exports" style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <h3 style={{ margin: 0 }}>Exports</h3>
        <div className="row">
          <button onClick={exportDrafts} disabled={busy}>
            Download ticket drafts (json/md/csv)
          </button>
          <small style={{ marginLeft: 'auto' }}>{(analysis?.ticket_drafts ?? []).length} drafts</small>
        </div>

        <small>
          Motion create button is handled in T-010; once enabled, we can add “Create in Motion” actions per draft here.
        </small>
      </section>
    </section>
  );
}
