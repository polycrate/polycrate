#!/usr/bin/env bash

# Copyright Ayedo Cloud Solutions GmbH.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# The install script is based off of the MIT-licensed script from glide,
# the package manager for Go: https://github.com/Masterminds/glide.sh/blob/master/get

: ${BINARY_NAME:="polycrate"}
: ${USE_SUDO:="true"}
: ${DEBUG:="false"}
: ${POLYCRATE_INSTALL_DIR:="/usr/local/bin"}

HAS_CURL="$(type "curl" &> /dev/null && echo true || echo false)"
HAS_WGET="$(type "wget" &> /dev/null && echo true || echo false)"

# initArch discovers the architecture for this system.
initArch() {
  ARCH=$(uname -m)
  case $ARCH in
    armv5*) ARCH="armv5";;
    armv6*) ARCH="armv6";;
    armv7*) ARCH="arm";;
    aarch64) ARCH="arm64";;
    x86) ARCH="386";;
    x86_64) ARCH="amd64";;
    i686) ARCH="386";;
    i386) ARCH="386";;
  esac
}

# initOS discovers the operating system for this system.
initOS() {
  OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')

  case "$OS" in
    # Minimalist GNU for Windows
    mingw*) OS='windows';;
  esac
}

# runs the given command as root (detects if we are root already)
runAsRoot() {
  if [ $EUID -ne 0 -a "$USE_SUDO" = "true" ]; then
    sudo "${@}"
  else
    "${@}"
  fi
}

# verifySupported checks that the os/arch combination is supported for
# binary builds, as well whether or not necessary tools are present.
verifySupported() {
  local supported="darwin-amd64\ndarwin-arm64\nlinux-386\nlinux-amd64\nlinux-arm64"
  if ! echo "${supported}" | grep -q "${OS}-${ARCH}"; then
    echo "No prebuilt binary for ${OS}-${ARCH}."
    exit 1
  fi

  if [ "${HAS_CURL}" != "true" ] && [ "${HAS_WGET}" != "true" ]; then
    echo "Either curl or wget is required"
    exit 1
  fi
}

get_latest_release() {
  curl --silent "https://hub.polycrate.io/api/v1/cli/get/?out=latest&type=plain"
  # curl --silent "https://api.github.com/repos/polycrate/polycrate/releases/latest" | # Get latest release from GitHub api
  #   grep '"tag_name":' |                                            # Get tag line
  #   sed -E 's/.*"([^"]+)".*/\1/'   |                              # Pluck JSON value
  #   cut -c 2-
}

# checkDesiredVersion checks if the desired version is available.
checkDesiredVersion() {
  if [ "x$DESIRED_VERSION" == "x" ]; then
    TAG=$(get_latest_release)
    echo "Latest version is $TAG"
  else
    TAG=$DESIRED_VERSION
  fi
}

# checkPolycrateInstalledVersion checks which version of Polycrate is installed and
# if it needs to be changed.
checkPolycrateInstalledVersion() {
  if [[ -f "${POLYCRATE_INSTALL_DIR}/${BINARY_NAME}" ]]; then
    local version=$("${POLYCRATE_INSTALL_DIR}/${BINARY_NAME}" version --short)
    if [[ "$version" == "$TAG" ]]; then
      echo "Polycrate ${version} is already ${DESIRED_VERSION:-latest}"
      return 0
    else
      echo "Polycrate ${TAG} is available. Changing from version ${version}."
      return 1
    fi
  else
    return 1
  fi
}

# downloadFile downloads the latest binary package and also the checksum
# for that binary.
downloadFile() {
  #DOWNLOAD_URL="https://s3.ayedo.de/polycrate/cli/v${TAG}/${BINARY_NAME}_${TAG}_${OS}_${ARCH}.tar.gz"
  DOWNLOAD_URL="https://hub.polycrate.io/get/polycrate/${TAG}/${OS}_${ARCH}/${BINARY_NAME}_${TAG}_${OS}_${ARCH}.tar.gz"
  # DOWNLOAD_URL="https://github.com/polycrate/polycrate/releases/download/v${TAG}/${BINARY_NAME}_${TAG}_${OS}_${ARCH}.tar.gz"

  POLYCRATE_TMP_ROOT="$(mktemp -dt polycrate-installer-XXXXXX)"
  POLYCRATE_TMP_FILE="$POLYCRATE_TMP_ROOT/${BINARY_NAME}.tar.gz"
  
  echo "Downloading $DOWNLOAD_URL"

  if [ "${HAS_CURL}" == "true" ]; then
    curl -SsL "$DOWNLOAD_URL" -o "$POLYCRATE_TMP_FILE"
  elif [ "${HAS_WGET}" == "true" ]; then
    wget -q -O "$POLYCRATE_TMP_FILE" "$DOWNLOAD_URL"
  fi
}

# installFile installs the Polycrate binary.
installFile() {
  POLYCRATE_TMP_BIN="$POLYCRATE_TMP_ROOT/$BINARY_NAME"
  echo "Unpacking $BINARY_NAME"

  tar xzf ${POLYCRATE_TMP_FILE} -C $POLYCRATE_TMP_ROOT

  if [ "$OS" == "WINDOWS" ] || [ "$OS" == "windows" ]; then
    BINARY_NAME=$BINARY_NAME.exe
  fi

  echo "Preparing to install $BINARY_NAME into ${POLYCRATE_INSTALL_DIR}"
  
  if [ -f "$POLYCRATE_INSTALL_DIR/$BINARY_NAME" ]; then
    runAsRoot rm "$POLYCRATE_INSTALL_DIR/$BINARY_NAME"
  fi
  runAsRoot cp "$POLYCRATE_TMP_BIN" "$POLYCRATE_INSTALL_DIR/$BINARY_NAME"
  runAsRoot chmod +x "$POLYCRATE_INSTALL_DIR/$BINARY_NAME"
  echo "$BINARY_NAME installed into $POLYCRATE_INSTALL_DIR/$BINARY_NAME"
}

# fail_trap is executed if an error occurs.
fail_trap() {
  result=$?
  if [ "$result" != "0" ]; then
    if [[ -n "$INPUT_ARGUMENTS" ]]; then
      echo "Failed to install $BINARY_NAME with the arguments provided: $INPUT_ARGUMENTS"
      help
    else
      echo "Failed to install $BINARY_NAME"
    fi
    echo -e "\tFor support, go to https://docs.ayedo.cloud/s/main/"
  fi
  cleanup
  exit $result
}

# testVersion tests the installed client to make sure it is working.
testVersion() {
  set +e
  POLYCRATE="$(command -v $BINARY_NAME)"
  if [ "$?" = "1" ]; then
    echo "$BINARY_NAME not found. Is $POLYCRATE_INSTALL_DIR on your "'$PATH?'
    exit 1
  fi
  set -e
}

# help provides possible cli installation arguments
help () {
  echo "Accepted cli arguments are:"
  echo -e "\t[--help|-h ] ->> prints this help"
  echo -e "\t[--version|-v <desired_version>] . When not defined it fetches the latest release from GitHub"
  echo -e "\te.g. --version v3.0.0 or -v canary"
  echo -e "\t[--no-sudo]  ->> install without sudo"
}

cleanup() {
  if [[ -d "${POLYCRATE_TMP_ROOT:-}" ]]; then
    rm -rf "$POLYCRATE_TMP_ROOT"
  fi
}

# Execution

#Stop execution on any error
trap "fail_trap" EXIT
set -e

# Set debug if desired
if [ "${DEBUG}" == "true" ]; then
  set -x
fi

# Parsing input arguments (if any)
export INPUT_ARGUMENTS="${@}"
set -u
while [[ $# -gt 0 ]]; do
  case $1 in
    '--version'|-v)
       shift
       if [[ $# -ne 0 ]]; then
           export DESIRED_VERSION="${1}"
       else
           echo -e "Please provide the desired version. e.g. --version 3.0.0 or -v latest"
           exit 0
       fi
       ;;
    '--no-sudo')
       USE_SUDO="false"
       ;;
    '--help'|-h)
       help
       exit 0
       ;;
    *) exit 1
       ;;
  esac
  shift
done
set +u

initArch
initOS
verifySupported
checkDesiredVersion
if ! checkPolycrateInstalledVersion; then
  downloadFile
  installFile
fi
testVersion
cleanup