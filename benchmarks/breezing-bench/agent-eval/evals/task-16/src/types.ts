export type MergeStrategy = 'replace' | 'append' | 'prefer-base';

export interface ConfigValue {
  [key: string]: string | number | boolean | string[] | ConfigValue;
}

export interface IConfigMerger {
  merge(base: ConfigValue, override: ConfigValue): ConfigValue;
  mergeWithStrategy(base: ConfigValue, override: ConfigValue, strategy: MergeStrategy): ConfigValue;
}
