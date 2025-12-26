# Deprecation Policy

[繁體中文](TW/deprecation.md)

## Policy

Our goal is to maintain stability for consumers. However, features may be deprecated to evolve the platform.

1. **Notice Period**: We pledge to provide at least **6 months** notice before removing a public API or feature.
2. **Communication**: Deprecations will be announced in Release Notes and marked in documentation.
3. **Runtime Warning**: Deprecated APIs will return a `Warning` HTTP header if possible.

### Deprecation Lifecycle

1. **Draft**: Internal discussion.
2. **Deprecated**: Marked as deprecated in docs. Fully functional. `Warning` header added.
3. **End of Support (EoS)**: Feature may still work but no bug fixes provided.
4. **Removed**: Feature is completely removed. Endpoint returns `404` or `410 Gone`.

### Current Deprecations

*None at the moment.*
