import typer
from loguru import logger
from rich import print
from ansible.inventory.manager import InventoryManager
from ansible.parsing.dataloader import DataLoader
from ansible.vars.manager import VariableManager
from types import SimpleNamespace
from typing import Optional
import utils
import yaml
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
def hosts(
        ctx: typer.Context,
        host: Optional[str] = typer.Argument("all"),
        output_file: str = typer.Option(""),
):
    hosts = {}
    try:
        load(ctx)
        #print(ctx.obj.ansible_inventory.get_hosts(pattern=host))
        for host in ctx.obj.ansible_inventory.get_hosts():
            host_vars = ctx.obj.ansible_vars.get_vars(host=host)

            # if ctx.obj.verbosity > 0:
            #     print(host_vars)
            # else:
            #     print(
            #         f"Host: {host_vars['inventory_hostname']}, IP: {host_vars['ansible_host']}"
            #     )
            hosts[host_vars['inventory_hostname']] = {}
            if "ansible_host" in host_vars:
                hosts[host_vars['inventory_hostname']]["ip"] = host_vars[
                    "ansible_host"]
            if "ansible_ssh_port" in host_vars:
                hosts[host_vars['inventory_hostname']]["port"] = host_vars[
                    'ansible_ssh_port']
            if "ansible_ssh_user" in host_vars:
                hosts[host_vars['inventory_hostname']]["user"] = host_vars[
                    'ansible_ssh_user']
            if "ansible_user" in host_vars:
                hosts[host_vars['inventory_hostname']]["user"] = host_vars[
                    'ansible_user']
        if output_file != "":
            with open(output_file, 'w') as file:
                utils.json_file(hosts, file)
        else:
            print(utils.yaml_dump(hosts))
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
