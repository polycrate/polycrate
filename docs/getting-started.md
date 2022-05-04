---
title: Getting Started
description: Install polycrate
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

You can download any version of polycrate from our [GitHub Releases](https://github.com/polycrate/polycrate/releases) by following the steps for your platform below:

=== "Linux"

    ``` bash
    curl -fsSLo polycrate.tar.gz https://github.com/polycrate/polycrate/releases/download/v0.2.2/polycrate_0.2.2_linux_amd64.tar.gz
    tar xvzf polycrate.tar.gz
    chmod +x polycrate
    ./polycrate version
    ```

=== "Linux (ARM)"

    ``` bash
    curl -fsSLo polycrate.tar.gz https://github.com/polycrate/polycrate/releases/download/v0.2.2/polycrate_0.2.2_linux_arm64.tar.gz
    tar xvzf polycrate.tar.gz
    chmod +x polycrate
    ./polycrate version
    ```

=== "macOS"

    ``` bash
    curl -fsSLo polycrate.tar.gz https://github.com/polycrate/polycrate/releases/download/v0.2.2/polycrate_0.2.2_darwin_amd64.tar.gz
    tar xvzf polycrate.tar.gz
    chmod +x polycrate
    ./polycrate version
    ```

=== "macOS (M1)"

    ``` bash
    curl -fsSLo polycrate.tar.gz https://github.com/polycrate/polycrate/releases/download/v0.2.2/polycrate_0.2.2_darwin_arm64.tar.gz
    tar xvzf polycrate.tar.gz
    chmod +x polycrate
    ./polycrate version
    ```

## Configuration

Cloudstack configuration is defined in a `Stackfile` in the context directory.

To see the default configuration, use `cloudstack show defaults`.