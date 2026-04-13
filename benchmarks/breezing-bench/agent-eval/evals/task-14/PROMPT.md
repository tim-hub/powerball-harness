Add a `getOrSet(key, factory, ttl?)` method to TTLCache.
The interface definition is in `types.ts`.
This feature returns the value if the key exists, otherwise calls the factory function, sets the result, and returns it.
The factory function should also support async functions.
