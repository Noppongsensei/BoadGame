import { NextResponse } from 'next/server';

export const dynamic = 'force-dynamic';

const resolveBackendBaseUrl = () => {
  const raw = process.env.API_URL || process.env.NEXT_PUBLIC_API_URL;
  if (!raw) return null;
  return raw.startsWith('http') ? raw : `https://${raw}`;
};

export async function GET() {
  const backendBaseUrl = resolveBackendBaseUrl();
  if (!backendBaseUrl) {
    return NextResponse.json(
      { error: 'API_URL is not configured on the frontend service' },
      { status: 500 }
    );
  }

  const u = new URL(backendBaseUrl);
  u.protocol = u.protocol === 'https:' ? 'wss:' : 'ws:';
  u.pathname = '/ws';
  u.search = '';

  return NextResponse.json({ ws_url: u.toString() });
}
