#!/bin/bash

set -e

#=======================================
# Functions
#=======================================

RESTORE='\033[0m'
RED='\033[00;31m'
YELLOW='\033[00;33m'
BLUE='\033[00;34m'
GREEN='\033[00;32m'

function color_echo {
	color=$1
	msg=$2
	echo -e "${color}${msg}${RESTORE}"
}

function echo_fail {
	msg=$1
	echo
	color_echo "${RED}" "${msg}"
	exit 1
}

function echo_warn {
	msg=$1
	color_echo "${YELLOW}" "${msg}"
}

function echo_info {
	msg=$1
	echo
	color_echo "${BLUE}" "${msg}"
}

function echo_details {
	msg=$1
	echo "  ${msg}"
}

function echo_done {
	msg=$1
	color_echo "${GREEN}" "  ${msg}"
}

function validate_required_input {
	key=$1
	value=$2
	if [ -z "${value}" ] ; then
		echo_fail "[!] Missing required input: ${key}"
	fi
}

function print_and_run {
  cmd="$1"
  echo_details "${cmd}"
	echo
  eval "${cmd}"
}

#=======================================
# Main
#=======================================

# Parameters
echo_info "Configs:"
echo_details "* source: $source"
echo_details "* destination: $destination"

validate_required_input "source" $source
validate_required_input "destination" $destination

# Download
echo_info "Downloading $source to $destination"

dir=$(dirname "${destination}")
dir="${dir/#\~/$HOME}"
base=$(basename "${destination}")

if [ -n "$dir" ] && [ ! -d "$dir" ] ; then
	mkdir -p "$dir"
fi

destination="$dir/$base"

set -x
wget "${source}" --output-document="${destination}"
