import type { ValidationResult, FormData, ValidationRule, IFormValidator } from './types';

export class FormValidator implements IFormValidator {
  private rules: ValidationRule[] = [];

  addRule(rule: ValidationRule): void {
    this.rules.push(rule);
  }

  validate(data: FormData): { isValid: boolean; errors: Record<string, string[]> } {
    const errors: Record<string, string[]> = {};

    for (const rule of this.rules) {
      const value = data[rule.field] ?? '';
      const fieldErrors: string[] = [];

      if (rule.required && !value.trim()) {
        fieldErrors.push(`${rule.field} is required`);
      }
      if (rule.minLength && value.length < rule.minLength) {
        fieldErrors.push(`${rule.field} must be at least ${rule.minLength} characters`);
      }
      if (rule.maxLength && value.length > rule.maxLength) {
        fieldErrors.push(`${rule.field} must be at most ${rule.maxLength} characters`);
      }
      if (rule.pattern && !rule.pattern.test(value)) {
        fieldErrors.push(`${rule.field} format is invalid`);
      }

      if (fieldErrors.length > 0) {
        errors[rule.field] = fieldErrors;
      }
    }

    const allErrors = Object.values(errors).flat();
    return {
      isValid: allErrors.length > 0,
      errors,
    };
  }

}
