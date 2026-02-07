export type HelperFn = (arg: string) => string;

export interface TemplateContext {
  [key: string]: string | number | boolean;
}

export interface ITemplateEngine {
  render(template: string, context: TemplateContext): string;
  registerHelper(name: string, fn: HelperFn): void;
}
