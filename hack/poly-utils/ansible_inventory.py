import typer
from loguru import logger
from rich import print
from ansible.inventory.manager import InventoryManager
from ansible.parsing.dataloader import DataLoader
from ansible.vars.manager import VariableManager
from types import SimpleNamespace
from typing import Optional
import utils
from ansible.module_utils.common.json import json_dump

app = typer.Typer()

# @app.command()
# def show(ctx: typer.Context, format: Optional[str] = typer.Option("yaml")):
#     help(ctx.obj.inventory)
#     if format == "yaml":
#         print(utils.yaml_dump(ctx.obj.inventory.get_hosts('all')))
#     else:
#         print(json_dump(ctx.obj.inventory))


@app.command()
def hosts(ctx: typer.Context, host: Optional[str] = typer.Argument("all")):
    print(ctx.obj.inventory.get_hosts(pattern=host))


@app.command()
def groups(ctx: typer.Context, group: Optional[str] = typer.Argument("all")):
    print(ctx.obj.inventory.get_groups_dict())


@app.command()
def list(ctx: typer.Context):
    for host in ctx.obj.inventory.get_hosts():
        host_vars = ctx.obj.vars.get_vars(host=host)

        if ctx.obj.verbosity > 0:
            print(host_vars)
        else:
            print(
                f"Host: {host_vars['inventory_hostname']}, IP: {host_vars['ansible_host']}"
            )


@app.command()
def halp(ctx: typer.Context):
    help(ctx.obj.inventory)
