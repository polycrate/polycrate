# Copyright 2017 RedHat, inc
#
# This file is part of Ansible
#
# Ansible is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# Ansible is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with Ansible.  If not, see <http://www.gnu.org/licenses/>.
#############################################
from __future__ import absolute_import, division, print_function

__metaclass__ = type

DOCUMENTATION = """
    name: polycrate_vars
    version_added: "2.9"
    short_description: In charge of loading variables from a polycrate workspace snapshot
    requirements:
        - Enabled in configuration
    description:
        - Loads YAML vars as extra vars from the workspace snapshot specified in POLYCRATE_WORKSPACE_SNAPSHOT_YAML
    options:
      stage:
        ini:
          - key: stage
            section: vars_polycrate_vars
        env:
          - name: ANSIBLE_VARS_PLUGIN_STAGE
    extends_documentation_fragment:
      - vars_plugin_staging
"""

import os
from ansible import constants as C
from ansible.errors import AnsibleParserError
from ansible.module_utils._text import to_bytes, to_native, to_text
from ansible.plugins.vars import BaseVarsPlugin
from ansible.inventory.host import Host
from ansible.inventory.group import Group
from ansible.utils.vars import combine_vars


class VarsModule(BaseVarsPlugin):

    REQUIRES_WHITELIST = False
    REQUIRES_ENABLED = False

    def get_vars(self, loader, path, entities, cache=True):
        """parses the inventory file"""

        if not isinstance(entities, list):
            entities = [entities]

        super(VarsModule, self).get_vars(loader, path, entities)

        data = {}

        POLYCRATE_WORKSPACE_SNAPSHOT_YAML = os.getenv(
            "POLYCRATE_WORKSPACE_SNAPSHOT_YAML"
        )

        try:
            b_opath = os.path.realpath(to_bytes(POLYCRATE_WORKSPACE_SNAPSHOT_YAML))
            opath = to_text(b_opath)

            if os.path.exists(opath):
                new_data = loader.load_from_file(opath, cache=True, unsafe=True)
                if new_data:  # ignore empty files
                    data = combine_vars(data, new_data)
            else:
                raise AnsibleParserError(f"File not found: {opath}")

        except Exception as e:
            raise AnsibleParserError(to_native(e))
        return data
