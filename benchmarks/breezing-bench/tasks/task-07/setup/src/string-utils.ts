export function capitalize(str: string): string {
  if (!str) return '';
  return str.charAt(0).toUpperCase() + str.slice(1);
}

export function camelCase(str: string): string {
  return str.replace(/[-_\s]+(.)?/g, (_, c) => c ? c.toUpperCase() : '');
}

export function kebabCase(str: string): string {
  return str.replace(/([a-z])([A-Z])/g, '$1-$2').replace(/[\s_]+/g, '-').toLowerCase();
}

export function snakeCase(str: string): string {
  return str.replace(/([a-z])([A-Z])/g, '$1_$2').replace(/[\s-]+/g, '_').toLowerCase();
}

export function truncate(str: string, maxLength: number, suffix: string = '...'): string {
  if (str.length <= maxLength) return str;
  return str.slice(0, maxLength - suffix.length) + suffix;
}

export function countWords(str: string): number {
  return str.trim().split(/\s+/).filter(Boolean).length;
}

export function reverse(str: string): string {
  return [...str].reverse().join('');
}

export function isPalindrome(str: string): boolean {
  const cleaned = str.toLowerCase().replace(/[^a-z0-9]/g, '');
  return cleaned === [...cleaned].reverse().join('');
}

export function padStart(str: string, length: number, char: string = ' '): string {
  while (str.length < length) str = char + str;
  return str;
}

export function padEnd(str: string, length: number, char: string = ' '): string {
  while (str.length < length) str = str + char;
  return str;
}

export function repeat(str: string, count: number): string {
  if (count < 0) throw new Error('count must be non-negative');
  return str.repeat(count);
}

export function contains(str: string, search: string, caseSensitive: boolean = true): boolean {
  if (!caseSensitive) return str.toLowerCase().includes(search.toLowerCase());
  return str.includes(search);
}

export function slugify(str: string): string {
  return str.toLowerCase().trim().replace(/[^\w\s-]/g, '').replace(/[\s_]+/g, '-').replace(/^-+|-+$/g, '');
}

export function escapeHtml(str: string): string {
  const map: Record<string, string> = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#039;' };
  return str.replace(/[&<>"']/g, c => map[c]);
}

export function unescapeHtml(str: string): string {
  const map: Record<string, string> = { '&amp;': '&', '&lt;': '<', '&gt;': '>', '&quot;': '"', '&#039;': "'" };
  return str.replace(/&amp;|&lt;|&gt;|&quot;|&#039;/g, m => map[m]);
}

export function extractEmails(str: string): string[] {
  const regex = /[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}/g;
  return str.match(regex) || [];
}

export function extractUrls(str: string): string[] {
  const regex = /https?:\/\/[^\s<>]+/g;
  return str.match(regex) || [];
}

export function wordWrap(str: string, maxWidth: number): string {
  const words = str.split(' ');
  const lines: string[] = [];
  let currentLine = '';
  for (const word of words) {
    if (currentLine && (currentLine + ' ' + word).length > maxWidth) {
      lines.push(currentLine);
      currentLine = word;
    } else {
      currentLine = currentLine ? currentLine + ' ' + word : word;
    }
  }
  if (currentLine) lines.push(currentLine);
  return lines.join('\n');
}

export function removeExtraSpaces(str: string): string {
  return str.replace(/\s+/g, ' ').trim();
}

export function isValidEmail(str: string): boolean {
  return /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/.test(str);
}
