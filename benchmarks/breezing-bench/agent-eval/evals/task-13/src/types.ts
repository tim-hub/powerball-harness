export interface ParsedHeader {
  [key: string]: string | undefined;
}

export interface SetCookieAttributes {
  name: string;
  value: string;
  expires?: string;
  maxAge?: number;
  path?: string;
  domain?: string;
  secure?: boolean;
  httpOnly?: boolean;
  sameSite?: 'Strict' | 'Lax' | 'None';
}

export type ParseHeaderFn = (raw: string) => ParsedHeader;
export type ParseSetCookieFn = (header: string) => SetCookieAttributes;
