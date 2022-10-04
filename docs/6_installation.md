---
title: Installation
description: Install polycrate
---

# Getting Started

## Requirements

Polycrate requires access to a local [Docker](https://www.docker.com/products/docker-desktop/) installation.

## Installation

### Automated installer

Using the polycrate installer automates the process of downloading and moving the `polycrate` binary to your `$PATH`. The installer automatically detects your operating system and architecture.

=== "curl"

    **real** quick:

    ```bash
    curl https://docs.polycrate.io/get-polycrate.sh | bash
    polycrate version
    ```

    a bit safer:

    ```bash
    curl -fsSL -o get-polycrate.sh https://docs.polycrate.io/get-polycrate.sh
    chmod 0700 get-polycrate.sh
    # Optionally: less get-polycrate.sh
    ./get-polycrate.sh
    polycrate version
    ```

=== "wget"

     **real** quick:

    ```bash
    wget -qO- https://docs.polycrate.io/get-polycrate.sh | bash
    polycrate version
    ```

    a bit safer:

    ```bash
    wget -q -O get-polycrate.sh https://docs.polycrate.io/get-polycrate.sh
    chmod 0700 get-polycrate.sh
    # Optionally: less get-polycrate.sh
    ./get-polycrate.sh
    polycrate version
    ```

### Manual Download

You can download any version of polycrate directly by following the steps for your platform below:

=== "Linux"

    ``` bash
    export VERSION=0.7.5
    curl -fsSLo polycrate.tar.gz https://s3.ayedo.de/polycrate/cli/v$VERSION/polycrate_$VERSION_linux_amd64.tar.gz
    tar xvzf polycrate.tar.gz
    chmod +x polycrate
    ./polycrate version
    ```

=== "Linux (ARM)"

    ``` bash
    export VERSION=0.7.5
    curl -fsSLo polycrate.tar.gz https://s3.ayedo.de/polycrate/cli/v$VERSION/polycrate_$VERSION_linux_arm64.tar.gz
    tar xvzf polycrate.tar.gz
    chmod +x polycrate
    ./polycrate version
    ```

=== "macOS"

    ``` bash
    export VERSION=0.7.5
    curl -fsSLo polycrate.tar.gz https://s3.ayedo.de/polycrate/cli/v$VERSION/polycrate_$VERSION_darwin_amd64.tar.gz
    tar xvzf polycrate.tar.gz
    chmod +x polycrate
    ./polycrate version
    ```

=== "macOS (M1)"

    ``` bash
    export VERSION=0.7.5
    curl -fsSLo polycrate.tar.gz https://s3.ayedo.de/polycrate/cli/v$VERSION/polycrate_$VERSION_darwin_amd64.tar.gz
    tar xvzf polycrate.tar.gz
    chmod +x polycrate
    ./polycrate version
    ```
