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
  <a href="https://docs.polycrate.io" alt="Docs">
    <img src="https://api.netlify.com/api/v1/badges/67a4a921-cfbb-442d-ae7c-a2f9439a4001/deploy-status" alt="Docs" />
  </a>
</p>

Polycrate is a framework that lets you package, integrate and automate complex applications and infrastructure. With Polycrate you can bundle dependencies, tools, cofiguration and deployment logic for any kind of IT system in a single workspace and expose reusable actions that enable a streamlined DevOps workflow.

## What is Polycrate

If you're working with modern cloud native tooling you most likely know the pain of dependency- and tool-management, the mess of git-repositories to get Infrastructure-as-Code working and the leaking portability when it comes to working in a team. Polycrate helps you to glue all the command-line tools, configuration files, secrets and deployment scripts together, package it into a version-controlled workspace and provide a seemless way to expose well-defined actions that allow for easy replication and low-risk execution of common workflows.

Polycrate does this by wrapping logic in so called blocks - custom code that you can write in the language of your choice - and execute them inside a Docker container that provides well-integrated best-of-breed tooling of modern infrastructure development. These blocks can be shared through a standard OCI registry or git-repositories, making a workspace portable between any system that supports Docker.

You can share workspaces with your team or customers and make them use pre-defined actions that setup, change or destroy the systems defined in it with simple commands like `polycrate run docker install`.

## Why Polycrate

- Simple commands to execute complex, pre-configured deployment or installation logic
- No need to locally install abstract toolchains and conflicting dependencies when working with tools like Ansible, Kubernetes or Terraform
- No knowledge of the logic of single blocks is necessary to use the exposed actions
- Build complex but well-integrated systems based on a single configuration file
- No custom DSL or complex configuration structure to learn: Polycrate lets you build on your own terms with minimum constraints, giving you the ability to configure your workspace and blocks the way YOU need to
- Share the operational load of managing complex systems in production with your team
- Works with and improves existing tools - no need to rewrite existing code to work with Polycrate. Just make a block out of it and expose it as an action
- Use tools like Ansible to achieve idempotent deployments inside your blocks
- Share common logic by pushing blocks to an OCI-comptabible registry
- Share workspaces through git-repositories

## Play with polycrate

- [Installation](https://docs.polycrate.io/getting-started)
- [Quick start](https://docs.polycrate.io/getting-started)
- [Examples](https://docs.polycrate.io/examples)

