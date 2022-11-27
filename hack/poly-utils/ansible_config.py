import typer
from rich import print
from loguru import logger
from types import SimpleNamespace
from typing import Optional
import utils
from ansible.module_utils.common.json import json_dump

app = typer.Typer()


@app.command()
def show(ctx: typer.Context, format: Optional[str] = typer.Option("yaml")):
    if format == "yaml":
        print(utils.yaml_dump(ctx.obj.config))
    else:
        print(json_dump(ctx.obj.config))


@app.command()
def halp(ctx: typer.Context):
    help(ctx.obj.inventory)
