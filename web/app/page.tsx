import MessageDemo from '@/components/MessageDemo';

export default function HomePage() {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <header className="card">
        <h1 style={{ margin: 0 }}>TEAmate</h1>
        <p style={{ margin: '8px 0 0 0' }}>
          Frontend scaffolding (Next.js + TypeScript). This page exercises the Go API via a local proxy route.
        </p>
        <small>
          Tip: copy <code>web/.env.local.example</code> to <code>web/.env.local</code> for local defaults.
        </small>
      </header>

      <MessageDemo />

      <section className="card">
        <h2 style={{ marginTop: 0 }}>Next steps (T-011)</h2>
        <ul style={{ margin: 0, paddingLeft: 18 }}>
          <li>Agenda input + shared transcript textbox (WebSocket feed)</li>
          <li>Drift view + outcomes + exports</li>
        </ul>
      </section>
    </div>
  );
}
