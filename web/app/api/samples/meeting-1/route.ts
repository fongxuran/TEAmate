import { NextResponse } from 'next/server';
import { readFile } from 'node:fs/promises';
import path from 'node:path';

export async function GET() {
  try {
    const p = path.resolve(process.cwd(), '../docs/transcript/meeting 1.txt');
    const text = await readFile(p, 'utf8');
    return new NextResponse(text, {
      status: 200,
      headers: { 'Content-Type': 'text/plain; charset=utf-8' },
    });
  } catch (e) {
    const message = e instanceof Error ? e.message : 'failed to read sample';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}
