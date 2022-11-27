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


def load(ctx: typer.Context):
    ctx.obj.ansible_inventory = InventoryManager(
        loader=ctx.obj.ansible_dataloader,
        sources=[ctx.obj.ansible_inventory_path])

    ctx.obj.ansible_vars = VariableManager(loader=ctx.obj.ansible_dataloader,
                                           inventory=ctx.obj.ansible_inventory)
    ctx.obj.ansible_vars._extra_vars = ctx.obj.snapshot


@app.command()
def hosts(ctx: typer.Context, host: Optional[str] = typer.Argument("all")):
    try:
        load(ctx)
        print(ctx.obj.inventory.get_hosts(pattern=host))
    except:
        raise RuntimeError("Inventory error")


@app.command()
def groups(ctx: typer.Context):
    load(ctx)
    print(ctx.obj.inventory.get_groups_dict())


@app.command()
def list(ctx: typer.Context):
    try:
        load(ctx)
        for host in ctx.obj.inventory.get_hosts():
            host_vars = ctx.obj.vars.get_vars(host=host)

            if ctx.obj.verbosity > 0:
                print(host_vars)
            else:
                print(
                    f"Host: {host_vars['inventory_hostname']}, IP: {host_vars['ansible_host']}"
                )
    except:
        raise RuntimeError("Inventory error")


@app.command()
def halp(ctx: typer.Context):
    help(ctx.obj.inventory)
