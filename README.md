<p align="center">
    <img src="https://raw.githubusercontent.com/polycrate/polycrate/main/logo.svg?sanitize=true"
        height="75">
</p>

<p align="center">
  <a href="https://discord.gg/8cQZfXWeXP" alt="Discord">
    <img src="https://img.shields.io/discord/971467892447146057?logo=discord" alt="Discord" />
  </a>
  <a href="https://github.com/polycrate/polycrate/blob/main/LICENSE" alt="License">
    <img src="https://img.shields.io/github/license/polycrate/polycrate" alt="License" />
  </a>
  <a href="https://github.com/polycrate/polycrate/blob/main/go.mod" alt="Go version">
    <img src="https://img.shields.io/github/go-mod/go-version/polycrate/polycrate" alt="Go version" />
  </a>
  <a href="https://github.com/polycrate/polycrate/releases" alt="Releases">
    <img src="https://img.shields.io/github/v/release/polycrate/polycrate" alt="GReleases" />
  </a>
</p>

Polycrate is a framework to build platforms. A platform can be anything from a bash script to automate your daily tasks, to a full-blown Kubernetes deployment.

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
