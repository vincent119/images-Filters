# Extensibility Guide

[繁體中文](TW/extensibility.md)

## Adding New Filters

Filters are implemented in `internal/processor`. To add a new filter:

1. **Interface**: Ensure your image processor implements the operation (e.g., specific contrast algorithm).
2. **Registry**: Register the filter name and parameter parser in the filter chain logic.
3. **Docs**: Update `api.md` with the new filter name and usage.

### Custom Loaders

To enable fetching images from a new source (e.g., Google Cloud Storage, FTP):

1. **Implement Interface**: Create a struct implementing `loader.ImageLoader`.

   ```go
   type ImageLoader interface {
       Load(ctx context.Context, path string) ([]byte, error)
   }
   ```

2. **Register**: Add the loader to the loader factory in `main.go`.

3. **Config**: Add necessary configuration keys to `config.yaml`.
