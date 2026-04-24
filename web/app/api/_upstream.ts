import type { NextRequest } from 'next/server';

function getUpstreamBaseUrl(): string {
  return (
    process.env.API_BASE_URL ||
    process.env.NEXT_PUBLIC_API_BASE_URL ||
    // safe local default
    'http://localhost:8080'
  );
}

function getAuthHeader(): string | null {
  const disabled = String(process.env.API_AUTH_DISABLED || '').toLowerCase();
  if (disabled === '1' || disabled === 'true' || disabled === 'yes') return null;

  const user = process.env.API_BASIC_AUTH_USER || 'admin';
  const pass = process.env.API_BASIC_AUTH_PASS || 'password';

  // If both are blank, assume upstream does not require auth.
  if (!user && !pass) return null;

  return `Basic ${Buffer.from(`${user}:${pass}`, 'utf8').toString('base64')}`;
}

export async function proxyUpstream(req: NextRequest, path: string): Promise<Response> {
  const upstream = new URL(path, getUpstreamBaseUrl());

  // Preserve query string
  req.nextUrl.searchParams.forEach((value: string, key: string) => {
    upstream.searchParams.set(key, value);
  });

  const headers: Record<string, string> = {
    Accept: 'application/json',
  };

  const auth = getAuthHeader();
  if (auth) headers.Authorization = auth;

  const init: RequestInit = {
    method: req.method,
    headers,
    cache: 'no-store',
  };

  if (req.method !== 'GET' && req.method !== 'HEAD') {
    headers['Content-Type'] = req.headers.get('content-type') || 'application/json';
    init.body = await req.text();
  }

  return fetch(upstream, init);
}
