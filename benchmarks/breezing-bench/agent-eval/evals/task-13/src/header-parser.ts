import type { ParsedHeader, SetCookieAttributes } from './types';

export function parseHeaders(raw: string): ParsedHeader {
  const result: ParsedHeader = {};
  const lines = raw.split('\n');
  for (const line of lines) {
    const parts = line.split(':');
    if (parts.length < 2) continue;
    const key = parts[0].trim().toLowerCase();
    const val = parts[1].trim();
    result[key] = val === '' ? undefined : val;
  }
  return result;
}

export function getContentLength(headers: ParsedHeader): number | undefined {
  const val = headers['content-length'];
  if (val == undefined) return undefined;
  const parsed = parseInt(val);
  return parsed;
}
