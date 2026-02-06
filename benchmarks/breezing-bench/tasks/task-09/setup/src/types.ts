export interface CsvParseOptions {
  delimiter?: string;    // default ','
  hasHeader?: boolean;   // default true
  skipEmpty?: boolean;   // default true
  trimValues?: boolean;  // default true
}

export interface CsvParseResult {
  headers: string[];
  rows: Record<string, string>[];
  errors: CsvError[];
}

export interface CsvError {
  line: number;
  message: string;
}
