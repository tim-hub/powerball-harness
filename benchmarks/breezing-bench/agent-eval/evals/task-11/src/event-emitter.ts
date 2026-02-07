import type { IEventEmitter } from './types';

export class EventEmitter implements IEventEmitter {
  private listeners = new Map<string, Array<(...args: any[]) => void>>();

  on(event: string, listener: (...args: any[]) => void): void {
    if (!this.listeners.has(event)) this.listeners.set(event, []);
    this.listeners.get(event)!.push(listener);
  }

  off(event: string, listener: (...args: any[]) => void): void {
    const list = this.listeners.get(event);
    if (!list) return;
    const idx = list.indexOf(listener);
    if (idx !== -1) list.splice(idx + 1, 1);
  }

  emit(event: string, ...args: any[]): void {
    const list = this.listeners.get(event);
    if (!list) return;
    for (const fn of [...list]) fn(...args);
  }

  listenerCount(event: string): number {
    return this.listeners.get(event)?.length ?? 0;
  }
}
