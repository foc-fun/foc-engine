#!/bin/bash
#
# Deploy & setup devnet contracts

FOC_ENGINE_CONFIG_FILE=$1
if [ -z "$FOC_ENGINE_CONFIG_FILE" ]; then
  echo "Usage: $0 <foc_engine_config_file> <project_script_dir> <foc_registry_contract_address>"
  exit 1
fi

PROJECT_SCRIPT_DIR=$2
if [ -z "$PROJECT_SCRIPT_DIR" ]; then
  echo "Usage: $0 <foc_engine_config_file> <project_script_dir> <foc_registry_contract_address>"
  exit 1
fi

FOC_REGISTRY_CONTRACT_ADDRESS=$3
if [ -z "$FOC_REGISTRY_CONTRACT_ADDRESS" ]; then
  echo "Usage: $0 <foc_engine_config_file> <project_script_dir> <foc_registry_contract_address>"
  exit 1
fi

ENV_FILE=$4

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

ETH_ADDRESS=0x49D36570D4E46F48E99674BD3FCC84644DDD6B96F7C741B1562B82F9E004DC7
DEVNET_ACCOUNT_ADDRESS=0x064b48806902a367c8598f4f95c305e8c1a1acba5f082d294a43793113115691
DEVNET_ACCOUNT_NAME="account-1"
DEVNET_ACCOUNT_FILE=oz_acct.json

RPC_HOST="localhost"
RPC_PORT=5050
RPC_URL=http://$RPC_HOST:$RPC_PORT

OUTPUT_DIR=$HOME/.foc-tests
TIMESTAMP=$(date +%s)
LOG_DIR=$OUTPUT_DIR/logs/$TIMESTAMP
TMP_DIR=$OUTPUT_DIR/tmp/$TIMESTAMP

# TODO: Clean option to remove old logs and state
#rm -rf $OUTPUT_DIR/logs/*
#rm -rf $OUTPUT_DIR/tmp/*
mkdir -p $LOG_DIR
mkdir -p $TMP_DIR

APPS=$(cat $FOC_ENGINE_CONFIG_FILE | jq -r '.contracts[]')
APP_COUNT=$(cat $FOC_ENGINE_CONFIG_FILE | jq -r '.contracts | length')
echo "Found $APP_COUNT apps in $FOC_ENGINE_CONFIG_FILE"
if [ "$APP_COUNT" -eq 0 ]; then
  echo "No apps found in $FOC_ENGINE_CONFIG_FILE"
  exit 0
fi

echo "Running init commands for each app"
for entry in $(echo $APPS | jq -r '. | @base64'); do
  _jq() {
    echo ${entry} | base64 --decode | jq -r ${1}
  }
  APP_NAME=$(_jq '.name')
  APP_CLASS_NAME=$(_jq '.class_name')
  APP_SETUP_COMMANDS=$(_jq '.init_commands[]')
  if [ -z "$APP_SETUP_COMMANDS" ]; then
    echo "No setup commands found for $APP_NAME"
    continue
  fi
  echo "Found setup commands for $APP_NAME: $APP_SETUP_COMMANDS"
  APP_SETUP_COMMANDS_COUNT=$(_jq -r '.init_commands | length')
  echo "1Found $APP_SETUP_COMMANDS_COUNT setup commands for $APP_NAME"
  if [ "$APP_SETUP_COMMANDS_COUNT" -eq 0 ]; then
    echo "No setup commands found for $APP_NAME"
    continue
  fi
  echo "2Found $APP_SETUP_COMMANDS_COUNT setup commands for $APP_NAME"
  APP_CONSTRUCTOR_ARGS_RAW=$(_jq '.constructor_args[]' | tr '\n' ' ')
  APP_CONSTRUCTOR_ARGS=$(echo $APP_CONSTRUCTOR_ARGS_RAW | sed "s/\$ACCOUNT/$DEVNET_ACCOUNT_ADDRESS/g")
  # TODO: Other replacements
  
  # Run setup commands
  echo "Processing app: $entry"
  for command in $APP_SETUP_COMMANDS; do
    echo "Running $APP_NAME setup command: $PROJECT_SCRIPT_DIR/$command"
    if [ -f "$PROJECT_SCRIPT_DIR/$command" ]; then
      APP_NAME=$APP_NAME \
        APP_CLASS_NAME=$APP_CLASS_NAME \
        APP_CONSTRUCTOR_ARGS="$APP_CONSTRUCTOR_ARGS" \
        FOC_REGISTRY_CONTRACT_ADDRESS=$FOC_REGISTRY_CONTRACT_ADDRESS \
        RPC_URL=$RPC_URL \
        LOG_DIR=$LOG_DIR \
        TMP_DIR=$TMP_DIR \
        ACCOUNT_FILE=$DEVNET_ACCOUNT_FILE \
        ACCOUNT_ADDRESS=$DEVNET_ACCOUNT_ADDRESS \
        ACCOUNT_NAME=$DEVNET_ACCOUNT_NAME \
        FOC_ENV_FILE=$ENV_FILE \
        bash $PROJECT_SCRIPT_DIR/$command
    else
      echo "Setup command file not found: $PROJECT_SCRIPT_DIR/$command"
    fi
  done
done
