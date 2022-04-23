# Configuration

Cloudstack configuration is defined in a `Stackfile` in the context directory.

To see the default configuration, use `cloudstack show defaults`.

To see the runtime configuration, use `cloudstack show config`.

## Troubleshooting

### Legacy configuration options

- `stack.mail`: is now `plugins.letsencrypt.config.mail`
