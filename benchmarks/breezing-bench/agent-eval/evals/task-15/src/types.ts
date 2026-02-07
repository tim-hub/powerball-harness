export interface ValidationResult {
  valid: boolean;
  error?: string;
}

export interface FormData {
  [field: string]: string;
}

export interface ValidationRule {
  field: string;
  required?: boolean;
  minLength?: number;
  maxLength?: number;
  pattern?: RegExp;
}

export interface IFormValidator {
  addRule(rule: ValidationRule): void;
  validate(data: FormData): { isValid: boolean; errors: Record<string, string[]> };
  validateEmail(email: string): ValidationResult;
  validateUrl(url: string): ValidationResult;
}
