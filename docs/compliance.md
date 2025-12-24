# Compliance & Privacy

[繁體中文](TW/compliance.md)

## Privacy

Images Filters is a data processor. It processes images provided by the client.

- **Data Retention**: Images are temporarily cached (Redis/Memory) for performance. We do not permanently store images unless configured with a persistent local/S3 storage backend.
- **Access Logs**: Web access logs (IP, User-Agent, URL) are generated for monitoring.
  - **GDPR**: IP addresses in logs may be considered PII.
  - **Recommendation**: Configure log rotation and retention policies (e.g., 30 days) in your deployment environment.

### Logging

Sensitive information (Parameters, Headers) should be sanitized before logging.

- **Security Keys**: Never logged.
- **Signatures**: Logged as part of URL.
