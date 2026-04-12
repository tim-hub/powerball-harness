Add a `getOrSet(key, factory, ttl?)` method to the TTLCache.
The interface definition is in `types.ts`.
This feature returns the value if the key exists, otherwise calls the factory function, sets the result, and returns it.
The factory should support async functions.
