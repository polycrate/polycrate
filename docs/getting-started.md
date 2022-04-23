---
title: Getting Started
description: Install the Cloudstack
---

# Getting Started

## Installation

### Automated installer

Using the [Cloudstack installer](Installer.md) automates the process of downloading and moving the `cloudstack` binary to your `$PATH`. The installer automatically detects your operating system and architecture.

=== "curl"

    **real** quick:

    ```bash
    curl https://docs.cloudstack.one/get-cloudstack.sh | bash
    ```

    a bit safer:

    ```bash
    curl -fsSL -o get-cloudstack.sh https://docs.cloudstack.one/get-cloudstack.sh
    chmod 0700 get-cloudstack.sh
    # Optionally: less get-cloudstack.sh
    ./get-cloudstack.sh
    ```

=== "wget"

     **real** quick:

    ```bash
    wget -qO- https://docs.cloudstack.one/get-cloudstack.sh | bash
    ```

    a bit safer:

    ```bash
    wget -q -O get-cloudstack.sh https://docs.cloudstack.one/get-cloudstack.sh
    chmod 0700 get-cloudstack.sh
    # Optionally: less get-cloudstack.sh
    ./get-cloudstack.sh
    ```

### Manual Download

Install the [CLI](https://gitlab.com/ayedocloudsolutions/cloudstack/cli) by following the steps for your platform below:

=== "Linux"

    ``` bash
    curl -fsSLo ./cloudstack https://s3.ayedo.dev/packages/cloudstack/latest/cloudstack-linux-amd64
    chmod +x cloudstack
    ./cloudstack version
    ```

=== "Linux (ARM)"

    ``` bash
    curl -fsSLo ./cloudstack https://s3.ayedo.dev/packages/cloudstack/latest/cloudstack-linux-arm64
    chmod +x cloudstack
    ./cloudstack version
    ```

=== "macOS"

    ``` bash
    curl -fsSLo ./cloudstack https://s3.ayedo.dev/packages/cloudstack/latest/cloudstack-darwin-amd64
    chmod +x cloudstack
    ./cloudstack version
    ```

=== "macOS (M1)"

    ``` bash
    curl -fsSLo ./cloudstack https://s3.ayedo.dev/packages/cloudstack/latest/cloudstack-darwin-arm64
    chmod +x cloudstack
    ./cloudstack version
    ```

## Configuration

Cloudstack configuration is defined in a `Stackfile` in the context directory.

To see the default configuration, use `cloudstack show defaults`.