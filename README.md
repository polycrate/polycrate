# Polycrate

## Development

You need:

- Go
- Docker
- [goreleaser](https://goreleaser.com/quick-start/)
- [svu](https://github.com/caarlos0/svu)
- [upx](https://upx.github.io/)

### Local testing

Run `make snapshot` - this will create a `dist` dir that contains the bundled artifacts. The command also builds the required docker images locally, but doesn't push them

### Release a new version

- Make sure the workspace is clean: `git status`
- Once the workspace is clean, run `make next` to see the next computed version