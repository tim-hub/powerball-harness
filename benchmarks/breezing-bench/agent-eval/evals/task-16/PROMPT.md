Add a `mergeWithStrategy(base, override, strategy)` method to the ConfigMerger.
The interface definition is in `types.ts`.
The strategy can be 'replace' | 'append' | 'prefer-base', controlling the array merge strategy for nested objects.
