import type { ConfigValue, MergeStrategy, IConfigMerger } from './types';

export class ConfigMerger implements IConfigMerger {
  merge(base: ConfigValue, override: ConfigValue): ConfigValue {
    return this.deepMerge(base, override);
  }

  private deepMerge(target: ConfigValue, source: ConfigValue): ConfigValue {
    for (const key of Object.keys(source)) {
      const sourceVal = source[key];
      const targetVal = target[key];

      if (
        typeof sourceVal === 'object' &&
        sourceVal !== null &&
        !Array.isArray(sourceVal) &&
        typeof targetVal === 'object' &&
        targetVal !== null &&
        !Array.isArray(targetVal)
      ) {
        target[key] = this.deepMerge(targetVal as ConfigValue, sourceVal as ConfigValue);
      } else {
        target[key] = sourceVal;
      }
    }
    return target;
  }
}
