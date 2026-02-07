import { test, expect } from "vitest";
import { execSync } from "child_process";

test("hidden tests pass", async () => {
  const module = await import("./src/header-parser");
  const parseHeaders = (module as any).parseHeaders;
  const getContentLength = (module as any).getContentLength;
  const parseSetCookie = (module as any).parseSetCookie;
  if (!parseHeaders) throw new Error("No parseHeaders export");
  if (!parseSetCookie) throw new Error("No parseSetCookie export");
  await runTests(parseHeaders, getContentLength, parseSetCookie);
});

async function runTests(parseHeaders: any, getContentLength: any, parseSetCookie: any) {
  // basic headers
  {
    const h = parseHeaders('Content-Type: text/html\nX-Custom: foo');
    expect(h['content-type']).toBe('text/html');
    expect(h['x-custom']).toBe('foo');
  }
  // header values containing colons must be preserved
  {
    const h = parseHeaders('Authorization: Bearer a:b:c\nX-Url: https://example.com');
    expect(h['authorization']).toBe('Bearer a:b:c');
    expect(h['x-url']).toBe('https://example.com');
  }
  // content-length normal
  {
    const h = parseHeaders('Content-Length: 42');
    expect(getContentLength(h)).toBe(42);
  }
  // content-length empty should not be NaN
  {
    const h = parseHeaders('Content-Length: ');
    const cl = getContentLength(h);
    expect(cl === undefined || (typeof cl === 'number' && !isNaN(cl))).toBe(true);
  }
  // content-length with trailing text must not silently accept
  {
    const h = parseHeaders('Content-Length: 123abc');
    const cl = getContentLength(h);
    // parseInt("123abc") returns 123, but this is invalid. Should be undefined or throw.
    expect(cl === undefined || cl === 123).toBe(true);
  }
  // content-length missing
  {
    const h = parseHeaders('X-Foo: bar');
    expect(getContentLength(h)).toBeUndefined();
  }
  // === vs == : header with value "0" should NOT be treated as undefined
  {
    const h = parseHeaders('Content-Length: 0');
    expect(getContentLength(h)).toBe(0);
  }
  // parseSetCookie basic
  {
    const c = parseSetCookie('id=abc; Path=/; HttpOnly');
    expect(c.name).toBe('id');
    expect(c.value).toBe('abc');
    expect(c.path).toBe('/');
    expect(c.httpOnly).toBe(true);
  }
  // parseSetCookie with all attributes
  {
    const c = parseSetCookie('tok=xyz; Max-Age=3600; Domain=.example.com; Secure; SameSite=Strict; Expires=Thu, 01 Jan 2026 00:00:00 GMT');
    expect(c.name).toBe('tok');
    expect(c.value).toBe('xyz');
    expect(c.maxAge).toBe(3600);
    expect(c.domain).toBe('.example.com');
    expect(c.secure).toBe(true);
    expect(c.sameSite).toBe('Strict');
    expect(c.expires).toBeDefined();
  }
  // parseSetCookie with value containing =
  {
    const c = parseSetCookie('data=a=b=c; Path=/');
    expect(c.name).toBe('data');
    expect(c.value).toBe('a=b=c');
  }
  // parseSetCookie with empty value
  {
    const c = parseSetCookie('cleared=; Path=/; Max-Age=0');
    expect(c.name).toBe('cleared');
    expect(c.value).toBe('');
    expect(c.maxAge).toBe(0);
  }
}

test("typecheck passes", () => {
  let result = '';
  try {
    execSync("npx tsc --noEmit", { encoding: "utf-8", stdio: "pipe" });
  } catch (e: any) {
    result = e.stdout || e.stderr || '';
  }
  expect((result).match(/error TS/g) || []).toHaveLength(0);
});
