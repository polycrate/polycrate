#!/usr/bin/env python3
import typer
from loguru import logger
from rich import print
from ansible.inventory.manager import InventoryManager
from ansible.parsing.dataloader import DataLoader
from ansible.vars.manager import VariableManager

from types import SimpleNamespace
from typing import Optional
import inventory
import ssh
import utils
import config
from pathlib import Path


#@app.callback
def root(
    ctx: typer.Context,
    snapshot_path: str = typer.Option(
        None, envvar="POLYCRATE_WORKSPACE_SNAPSHOT_YAML"),
    inventory_path: str = typer.Option(None, envvar="ANSIBLE_INVENTORY"),
    ssh_user: str = typer.Option("root", envvar="ANSIBLE_SSH_USER"),
    ssh_private_key_file: str = typer.Option(
        f"{Path.home()}/.ssh/id_rsa", envvar="ANSIBLE_SSH_PRIVATE_KEY_FILE"),
    verbosity: int = typer.Option(0, envvar="ANSIBLE_VERBOSITY"),
    ssh_port: int = typer.Option(22, envvar="ANSIBLE_SSH_PORT"),
    #ipv4: bool = typer.Option(True, envvar="NS0_IPV4"),
):

    ctx.obj = SimpleNamespace()

    ctx.obj.inventory_path = inventory_path
    ctx.obj.verbosity = verbosity
    ctx.obj.ssh_port = ssh_port
    ctx.obj.ssh_user = ssh_user
    ctx.obj.ssh_private_key_file = ssh_private_key_file
    ctx.obj.config = utils.loadConfig()

    if inventory_path:
        dl = DataLoader()
        ctx.obj.inventory = InventoryManager(loader=dl,
                                             sources=[inventory_path])

        ctx.obj.vars = VariableManager(loader=dl, inventory=ctx.obj.inventory)
    else:
        print(f"[red]No inventory path[/red] given")
        raise typer.Exit(code=1)


app = typer.Typer(callback=root)
app.add_typer(inventory.app, name="inventory")
app.add_typer(ssh.app, name="ssh")
app.add_typer(config.app, name="config")

if __name__ == "__main__":
    app()