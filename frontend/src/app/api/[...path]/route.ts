import { NextResponse, type NextRequest } from 'next/server';

export const dynamic = 'force-dynamic';

const resolveBackendBaseUrl = () => {
  const raw = process.env.API_URL || process.env.NEXT_PUBLIC_API_URL;
  if (!raw) return null;
  return raw.startsWith('http') ? raw : `https://${raw}`;
};

const buildTargetUrl = (req: NextRequest) => {
  const backendBaseUrl = resolveBackendBaseUrl();
  if (!backendBaseUrl) {
    return null;
  }

  const reqUrl = new URL(req.url);
  const target = new URL(reqUrl.pathname + reqUrl.search, backendBaseUrl);
  return target;
};

const proxy = async (req: NextRequest) => {
  try {
    const targetUrl = buildTargetUrl(req);
    if (!targetUrl) {
      return NextResponse.json(
        {
          error: 'API_URL is not configured on the frontend service',
          hint: 'Set API_URL (recommended) or NEXT_PUBLIC_API_URL on the frontend service to point to the backend public URL',
        },
        { status: 500 }
      );
    }

    const headers = new Headers(req.headers);
    headers.delete('host');

    const init: RequestInit = {
      method: req.method,
      headers,
      redirect: 'manual',
      cache: 'no-store',
    };

    if (req.method !== 'GET' && req.method !== 'HEAD') {
      init.body = await req.arrayBuffer();
    }

    let upstream: Response;
    try {
      upstream = await fetch(targetUrl.toString(), init);
    } catch (e: any) {
      return NextResponse.json(
        {
          error: 'Failed to reach backend API',
          details: e?.message || String(e),
          target: targetUrl.toString(),
        },
        { status: 502 }
      );
    }

    const resHeaders = new Headers(upstream.headers);
    resHeaders.delete('content-encoding');

    const body = await upstream.arrayBuffer();
    return new NextResponse(body, {
      status: upstream.status,
      headers: resHeaders,
    });
  } catch (e: any) {
    return NextResponse.json(
      {
        error: 'Unhandled API proxy error',
        details: e?.message || String(e),
      },
      { status: 500 }
    );
  }
};

export async function GET(req: NextRequest) {
  return proxy(req);
}

export async function POST(req: NextRequest) {
  return proxy(req);
}

export async function PUT(req: NextRequest) {
  return proxy(req);
}

export async function PATCH(req: NextRequest) {
  return proxy(req);
}

export async function DELETE(req: NextRequest) {
  return proxy(req);
}

export async function OPTIONS(req: NextRequest) {
  return proxy(req);
}
