import type { TemplateContext, HelperFn, ITemplateEngine } from './types';

export class TemplateEngine implements ITemplateEngine {
  private helpers = new Map<string, HelperFn>();

  render(template: string, context: TemplateContext): string {
    let result = template.replace(/\{\{(\w+)\}\}/g, (_match, key) => {
      const value = context[key];
      if (value === undefined) return '';
      return String(value);
    });

    return result;
  }

  registerHelper(_name: string, _fn: HelperFn): void {
    // Not implemented yet — agent needs to add this
    throw new Error('Not implemented');
  }
}
