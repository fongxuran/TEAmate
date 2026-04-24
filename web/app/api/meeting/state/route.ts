import { NextResponse, type NextRequest } from 'next/server';

import { proxyUpstream } from '@/app/api/_upstream';

export async function GET(req: NextRequest) {
  try {
    const res = await proxyUpstream(req, '/api/meeting/state');
    const body = await res.text();
    return new NextResponse(body, {
      status: res.status,
      headers: {
        'Content-Type': res.headers.get('content-type') || 'application/json',
      },
    });
  } catch (e) {
    const message = e instanceof Error ? e.message : 'proxy failed';
    return NextResponse.json({ error: message }, { status: 502 });
  }
}
