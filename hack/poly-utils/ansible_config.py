import typer
from rich import print
from loguru import logger
from types import SimpleNamespace
from typing import Optional
import utils
#from ansible.module_utils.common.json import json_dump
from ansible import constants as C
from benedict import benedict
import yaml
from pathlib import Path
import os

app = typer.Typer()


def load():
    # Dump Env
    print(os.environ)
    # Load Ansible config
    config_entries = C.config.get_configuration_definitions(
        ignore_private=True)
    # for setting in ctx.obj.config.keys():
    #     v, o = C.config.get_config_value_and_origin(setting)
    #     ctx.obj.config[setting] = Setting(setting, v, o, None)
    return config_entries


def loadSnapshot(snapshot_path):
    try:
        # load yaml file
        snapshot = yaml.safe_load(Path(snapshot_path).read_text())
        #snapshot = benedict.from_yaml(snapshot_path, keypath_separator=None)

        return snapshot
    except Exception as e:
        raise RuntimeError(
            f"Unable to load snapshot from {snapshot_path}: {e}")


# @app.command()
# def show(ctx: typer.Context, format: Optional[str] = typer.Option("yaml")):
#     if format == "yaml":
#         print(utils.yaml_dump(ctx.obj.ansible_config))
#     else:
#         print(json_dump(ctx.obj.ansible_config))


@app.command()
def halp(ctx: typer.Context):
    help(ctx.obj.inventory)
