#!/bin/bash
#
# Run your app within the foc-engine

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
PROJECT_ROOT=$SCRIPT_DIR/..
ONCHAIN_SCRIPTS_DIR=$PROJECT_ROOT/onchain/scripts

# Set default values
FOC_ENGINE_URL="http://localhost:8085"

display_help() {
  echo "Usage: $0 [option...]"
  echo
  echo "   -p, --project <path>        path to the app's Starknet Scarb project ( Required )"
  echo "   -c, --config <path>         path to the app's config file ( default: <project>/foc-engine.config.json )"
  echo "   -e, --engine <url>          url to the foc-engine ( default: $FOC_ENGINE_URL )"
  echo "   -h, --help                  display help"

  echo
  echo "Example: $0"
}

# Transform long options to short ones
for arg in "$@"; do
  shift
  case "$arg" in
    "--help") set -- "$@" "-h" ;;
    "--project") set -- "$@" "-p" ;;
    "--config") set -- "$@" "-c" ;;
    "--engine") set -- "$@" "-e" ;;
    --*) unrecognized_options+=("$arg") ;;
    *) set -- "$@" "$arg"
  esac
done

# Check if unknown options are passed, if so exit
if [ ! -z "${unrecognized_options[@]}" ]; then
  echo "Error: invalid option(s) passed ${unrecognized_options[*]}"
  exit 1
fi

# Parse command line arguments
while getopts ":hp:c:e:" opt; do
  case ${opt} in
    h )
      display_help
      exit 0
      ;;
    p )
      APP_PROJECT_PATH=$OPTARG
      if [ ! -d "$APP_PROJECT_PATH" ]; then
        echo "Error: The provided path does not exist or is not a directory."
        exit 1
      fi
      ;;
    c )
      FOC_APP_CONFIG=$OPTARG
      if [ ! -f "$FOC_APP_CONFIG" ]; then
        echo "Error: The provided config file does not exist."
        exit 1
      fi
      ;;
    e )
      FOC_ENGINE_URL=$OPTARG
      ;;
    \? )
      echo "Invalid Option: -$OPTARG"
      display_help
      exit 1
      ;;
    : )
      echo "Invalid Option: -$OPTARG requires an argument"
      display_help
      exit 1
      ;;
  esac
done

# Check if the required argument is provided
if [ -z "$APP_PROJECT_PATH" ]; then
  echo "Error: The --project argument is required."
  display_help
  exit 1
fi

# Default values
if [ -z "$FOC_APP_CONFIG" ]; then
  FOC_APP_CONFIG=$APP_PROJECT_PATH/foc-engine.config.json
fi

DEFAULT_APP_SCRIPTS_DIR=$APP_PROJECT_PATH/scripts
GET_REGISTRY_CONTRACTS_RESULT=$(curl -s $FOC_ENGINE_URL/registry/get-registry-contracts)
FOC_REGISTRY_CONTRACT_ADDRESS=$(echo $GET_REGISTRY_CONTRACTS_RESULT | jq -r '.data.registry_contracts[-1]')

echo "Deploying app to FOC engine at $FOC_ENGINE_URL"
echo "$ONCHAIN_SCRIPTS_DIR/deploy_apps.sh $FOC_APP_CONFIG $APP_PROJECT_PATH $FOC_REGISTRY_CONTRACT_ADDRESS"
$ONCHAIN_SCRIPTS_DIR/deploy_apps.sh $FOC_APP_CONFIG $APP_PROJECT_PATH $FOC_REGISTRY_CONTRACT_ADDRESS

# TODO: Remove .foc_env file
echo
echo "Setting up app environment"
$ONCHAIN_SCRIPTS_DIR/setup_apps.sh $FOC_APP_CONFIG $DEFAULT_APP_SCRIPTS_DIR $FOC_REGISTRY_CONTRACT_ADDRESS .foc_env
