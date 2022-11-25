import typer
from loguru import logger
from rich import print
from ansible.inventory.manager import InventoryManager
from ansible.parsing.dataloader import DataLoader
from ansible.vars.manager import VariableManager
from types import SimpleNamespace
from typing import Optional
import subprocess

app = typer.Typer()


@app.command()
def exec(ctx: typer.Context):
    print("exec")


@app.command()
def shell(ctx: typer.Context, hostname: str = typer.Argument("")):
    host = ctx.obj.inventory.get_host(hostname)

    if not host:
        raise SystemExit(f"Host not found: {hostname}")

    host_vars = ctx.obj.vars.get_vars(host=host)
    user = ""

    if 'ansible_user' in host_vars.keys():
        user = host_vars['ansible_user']
    else:
        user = ctx.obj.ssh_user

    try:
        _v = ""
        if ctx.obj.verbosity > 0:
            _v = "-"
            _v = _v.ljust(ctx.obj.verbosity + len(_v), "v")

        ssh_cmd = f"ssh {_v} -i {ctx.obj.ssh_private_key_file} -o BatchMode=yes -p {ctx.obj.ssh_port} {user}@{host_vars['ansible_host']}"

        print(f"Executing ssh command: {ssh_cmd}")
        subprocess.run(ssh_cmd, shell=True)

    except subprocess.CalledProcessError as e:
        raise SystemExit(e)


if __name__ == "__main__":
    app()