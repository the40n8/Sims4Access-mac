// SimSight meeting point: one fixed address for the phone app, which then
// finds the Mac wherever its tunnel is today.
//
// - Serves the phone app (the same one the Mac serves), marked as hosted here.
// - POST /mp/announce: a Mac says where it is (its tunnel address) and,
//   while one is shown, its pairing code (hashed).
// - GET /mp/where?id=: the phone asks where its Mac is.
// - POST /mp/code: the phone finds which Mac a pairing code belongs to;
//   the Mac itself checks the code and gives the key.
//
// No game data goes through here, only addresses. A Mac's address alone is
// useless without its pairing key, which never comes here.
import { DurableObject } from 'cloudflare:workers';

const ONLINE_MS = 15 * 60 * 1000; // a Mac announces itself every 10 minutes
const FORGET_MS = 30 * 24 * 3600 * 1000;
const CODE_TRIES = 10; // per address and 10 minutes, all Macs together
const ID = /^[A-Za-z0-9_-]{20,64}$/;

export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    if (url.pathname.startsWith('/mp/')) return api(request, url, env);
    const res = await env.ASSETS.fetch(request);
    if (url.pathname === '/manifest.json') return manifest(request, res);
    if (!(res.headers.get('content-type') || '').includes('text/html')) return res;
    // The app: tell it it is hosted here, and let it reach the Macs' tunnels.
    return new HTMLRewriter()
      .on('meta[http-equiv="content-security-policy"]', {
        element(el) {
          const csp = el.getAttribute('content') || '';
          el.setAttribute('content', csp.replace("connect-src 'self'", "connect-src 'self' https:"));
        }
      })
      .on('head', {
        element(el) {
          el.append('<meta name="simsight-meeting-point" content="1">', { html: true });
        }
      })
      .transform(res);
  }
};

function json(body, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'content-type': 'application/json', 'cache-control': 'no-store' }
  });
}

async function sha256(text) {
  const d = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(text));
  return [...new Uint8Array(d)].map((b) => b.toString(16).padStart(2, '0')).join('');
}

async function body(request) {
  try {
    const text = await request.text();
    return text.length < 2048 ? JSON.parse(text) : null;
  } catch {
    return null;
  }
}

async function api(request, url, env) {
  const registry = env.REGISTRY.get(env.REGISTRY.idFromName('registry'));
  if (url.pathname === '/mp/hello') return json({ meetingPoint: true });
  if (url.pathname === '/mp/where' && request.method === 'GET') {
    const id = url.searchParams.get('id') || '';
    if (!ID.test(id)) return json({ error: 'bad id' }, 400);
    const mac = await registry.where(id);
    return mac ? json(mac) : json({ error: 'unknown' }, 404);
  }
  if (request.method !== 'POST') return json({ error: 'method' }, 405);
  const b = await body(request);
  if (!b) return json({ error: 'bad body' }, 400);
  if (url.pathname === '/mp/announce') {
    const { id, secret, name = '', url: link = '', code = '', codeSeconds = 0 } = b;
    if (!ID.test(id || '') || typeof secret !== 'string' || secret.length < 20) return json({ error: 'bad id' }, 400);
    if (link && !/^https:\/\/[^/\s]+$/.test(link)) return json({ error: 'bad url' }, 400);
    const ok = await registry.announce({
      id,
      secretHash: await sha256(secret),
      name: String(name).slice(0, 60),
      url: link,
      codeHash: /^\d{6}$/.test(code) ? await sha256(code) : '',
      codeUntil: Date.now() + Math.min(Number(codeSeconds) || 0, 600) * 1000
    });
    return ok ? json({ ok: true }) : json({ error: 'not this Mac' }, 403);
  }
  if (url.pathname === '/mp/code') {
    const code = String(b.code || '').replace(/\D/g, '');
    const ip = request.headers.get('cf-connecting-ip') || 'unknown';
    if (!(await registry.tryCode(ip))) return json({ error: 'too many tries' }, 429);
    if (code.length !== 6) return json({ error: 'unknown' }, 404);
    const mac = await registry.byCode(await sha256(code));
    return mac ? json(mac) : json({ error: 'unknown' }, 404);
  }
  return json({ error: 'not found' }, 404);
}

// The Home Screen app has its own storage on iOS: like the Mac's own server,
// put the phone's games in the install's start address (cookie set by the app).
async function manifest(request, res) {
  const cookie = request.headers.get('cookie') || '';
  const m = /(?:^|;\s*)ssg=([A-Za-z0-9_.~-]{20,2000})/.exec(cookie);
  if (!m || !res.ok) return res;
  const data = await res.json();
  data.start_url = './#g=' + m[1];
  return new Response(JSON.stringify(data), {
    headers: { 'content-type': 'application/manifest+json', 'cache-control': 'no-store' }
  });
}

export class Registry extends DurableObject {
  constructor(ctx, env) {
    super(ctx, env);
    this.sql = ctx.storage.sql;
    this.sql.exec(`CREATE TABLE IF NOT EXISTS macs (
      id TEXT PRIMARY KEY, secret_hash TEXT NOT NULL, name TEXT, url TEXT,
      seen INTEGER, code_hash TEXT, code_until INTEGER)`);
    this.sql.exec('CREATE TABLE IF NOT EXISTS tries (ip TEXT PRIMARY KEY, n INTEGER, since INTEGER)');
  }

  announce({ id, secretHash, name, url, codeHash, codeUntil }) {
    const now = Date.now();
    this.sql.exec('DELETE FROM macs WHERE seen < ?', now - FORGET_MS);
    const row = this.sql.exec('SELECT secret_hash FROM macs WHERE id = ?', id).toArray()[0];
    if (row && row.secret_hash !== secretHash) return false; // first come keeps the id
    this.sql.exec(
      `INSERT INTO macs (id, secret_hash, name, url, seen, code_hash, code_until) VALUES (?, ?, ?, ?, ?, ?, ?)
       ON CONFLICT(id) DO UPDATE SET name = excluded.name, url = excluded.url, seen = excluded.seen,
         code_hash = excluded.code_hash, code_until = excluded.code_until`,
      id, secretHash, name, url, now, codeHash, codeUntil
    );
    return true;
  }

  #public(row) {
    const online = !!row.url && Date.now() - row.seen < ONLINE_MS;
    return { id: row.id, name: row.name, url: online ? row.url : '', online, seen: row.seen };
  }

  where(id) {
    const row = this.sql.exec('SELECT * FROM macs WHERE id = ?', id).toArray()[0];
    return row ? this.#public(row) : null;
  }

  byCode(codeHash) {
    const row = this.sql
      .exec('SELECT * FROM macs WHERE code_hash = ? AND code_until > ? ORDER BY seen DESC LIMIT 1', codeHash, Date.now())
      .toArray()[0];
    return row ? this.#public(row) : null;
  }

  tryCode(ip) {
    const now = Date.now();
    const row = this.sql.exec('SELECT n, since FROM tries WHERE ip = ?', ip).toArray()[0];
    if (!row || now - row.since > 10 * 60 * 1000) {
      this.sql.exec('INSERT OR REPLACE INTO tries (ip, n, since) VALUES (?, 1, ?)', ip, now);
      this.sql.exec('DELETE FROM tries WHERE since < ?', now - 3600 * 1000);
      return true;
    }
    if (row.n >= CODE_TRIES) return false;
    this.sql.exec('UPDATE tries SET n = n + 1 WHERE ip = ?', ip);
    return true;
  }
}
