#!/usr/bin/env python3
import typer
from loguru import logger
from rich import print
from ansible.inventory.manager import InventoryManager
from ansible.parsing.dataloader import DataLoader
from ansible.vars.manager import VariableManager
from ansible import context
from ansible.cli import CLI
from types import SimpleNamespace
from typing import Optional
import yaml
from pathlib import Path
from benedict import benedict


import ansible_inventory
import ansible_config
import ssh
import utils
from pathlib import Path
import os
import sys


#@app.callback
def root(
    ctx: typer.Context,
    snapshot_path: str = typer.Option(
        None, envvar="POLYCRATE_WORKSPACE_SNAPSHOT_YAML"),
    inventory_path: str = typer.Option(None, envvar="ANSIBLE_INVENTORY"),
    ssh_user: str = typer.Option("root", envvar="USERNAME"),
    ssh_private_key_file: str = typer.Option(
        f"{Path.home()}/.ssh/id_rsa", envvar="ANSIBLE_SSH_PRIVATE_KEY_FILE"),
    verbosity: int = typer.Option(0, envvar="ANSIBLE_VERBOSITY"),
    ssh_port: int = typer.Option(22, envvar="ANSIBLE_SSH_PORT"),
    #ipv4: bool = typer.Option(True, envvar="NS0_IPV4"),
):
    logger.remove()

    match verbosity:
      case 1:
        logger.add(sys.stderr, level='INFO')
      case 2:
        logger.add(sys.stderr, level='DEBUG')
      case 3:  
        logger.add(sys.stderr, level='TRACE')
      case _:
        logger.add(sys.stderr, level='INFO')

    logger.debug(f"Running poly-utils")

    ctx.obj = SimpleNamespace()

    ctx.obj.ansible_inventory_path = inventory_path
    ctx.obj.snapshot_path = snapshot_path
    ctx.obj.verbosity = verbosity
    ctx.obj.ssh_port = ssh_port
    ctx.obj.ssh_user = ssh_user
    ctx.obj.ssh_private_key_file = ssh_private_key_file
    ctx.obj.ansible_config = ansible_config.load()
    ctx.obj.ansible_dataloader = DataLoader()

    logger.debug(f"Inventory path: {inventory_path}")
    logger.debug(f"Snapshot path: {snapshot_path}")
    # ctx.obj.inventory = InventoryManager(loader=ctx.obj.ansible_dataloader,
    #                       sources=[inventory_path])

    # ctx.obj.vars = VariableManager(loader=ctx.obj.ansible_dataloader, inventory=ctx.obj.inventory)
    
    ctx.obj.snapshot = ansible_config.loadSnapshot(snapshot_path)
    
    

app = typer.Typer(callback=root)
app.add_typer(ansible_inventory.app, name="ansible_inventory")
app.add_typer(ssh.app, name="ssh")
app.add_typer(ansible_config.app, name="ansible_config")

if __name__ == "__main__":
    app()