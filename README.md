# Polycrate

Polycrate is a framework to build platforms. Read more in the [official docs](https://docs.polycrate.io).

![Discord](https://img.shields.io/discord/971467892447146057)
![License](https://img.shields.io/github/license/polycrate/polycrate)

## Play with polycrate

- [Installation](https://docs.polycrate.io/getting-started)
- [Quick start](https://docs.polycrate.io/getting-started)
- [Examples](https://docs.polycrate.io/examples)

## Develop polycrate

You need:

- Go
- Docker
- [goreleaser](https://goreleaser.com/quick-start/)
- [svu](https://github.com/caarlos0/svu)
- [upx](https://upx.github.io/)

### Local testing

Run `make snapshot` - this will create a `dist` dir that contains the bundled artifacts. The command also builds the required docker images locally, but doesn't push them

### Release a new version

- Create a changelog and fill in necessary information for the release: `make changelog`
- Make sure the workspace is clean: `git status`
- Once the workspace is clean, run `make next` to see the next computed version
- If everything fits, run `make tag` - this will create and push a new tag
- Next, run `make release`

### Troubleshooting

#### error=git tag v0.2.0 was not made against commit $COMMIT

This happens if you pushed a tag and then made new changes before running `make release`. This can be solved running the following commands after cleaning the workspace:

- `make delete-tag`
- `make tag`
- `make release`
