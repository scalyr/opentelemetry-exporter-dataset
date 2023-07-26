#!/usr/bin/env bash

# This file contains common utility function and environment variables which are sourced and
# available to all the test scripts.

function log() {
  echo "$( gdate -Is 2> /dev/null || date -Is ) $*";
}

function enable_debug_mode_if_set() {
  DEBUG="${DEBUG:-"0"}"

  if [ "${DEBUG}" -eq "1" ]; then
    echo "DEBUG environment variable is set, enabling debug mode"
    set -x
  fi
}
