# Versioning Policy

[繁體中文](TW/versioning.md)

## API Versioning

Images Filters uses **URL Path Versioning** logic conceptually, although the current implementation exposes the root path.

**Future Roadmap:**

- V1: `http://host/v1/signature/...`
- V2: `http://host/v2/signature/...`

Currently, the service operates as **v1** implicitly. Any breaking change to the URL structure or signature algorithm will prompt a version bump to `v2`.

### Semantic Versioning

The application binary and Docker images follow [Semantic Versioning 2.0.0](https://semver.org/).

- **MAJOR**: Incompatible API changes.
- **MINOR**: Backward-compatible functionality (new filters, new loaders).
- **PATCH**: Backward-compatible bug fixes.
