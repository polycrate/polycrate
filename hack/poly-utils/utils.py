from ansible.parsing.yaml.dumper import AnsibleDumper
import yaml
from ansible import constants as C
from ansible.config.manager import ConfigManager, Setting
import typer


def yaml_dump(data, default_flow_style=False, default_style=None):
    return yaml.dump(data,
                     Dumper=AnsibleDumper,
                     default_flow_style=default_flow_style,
                     default_style=default_style)


def yaml_file(data, file, default_flow_style=False, default_style=None):
    yaml.dump(data,
              file,
              default_flow_style=default_flow_style,
              default_style=default_style)


def loadConfig():
    # Load Ansible config
    config_entries = C.config.get_configuration_definitions(
        ignore_private=True)
    # for setting in ctx.obj.config.keys():
    #     v, o = C.config.get_config_value_and_origin(setting)
    #     ctx.obj.config[setting] = Setting(setting, v, o, None)
    return config_entries
