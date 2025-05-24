#!/bin/bash
#
# Deploy & setup devnet contracts

FOC_ENGINE_CONFIG_FILE=$1
if [ -z "$FOC_ENGINE_CONFIG_FILE" ]; then
  echo "Usage: $0 <foc_engine_config_file> <project_dir> <foc_registry_contract_address>"
  exit 1
fi

PROJECT_DIR=$2
if [ -z "$PROJECT_DIR" ]; then
  echo "Usage: $0 <foc_engine_config_file> <project_dir> <foc_registry_contract_address>"
  exit 1
fi

FOC_REGISTRY_CONTRACT_ADDRESS=$3
if [ -z "$FOC_REGISTRY_CONTRACT_ADDRESS" ]; then
  echo "Usage: $0 <foc_engine_config_file> <project_dir> <foc_registry_contract_address>"
  exit 1
fi

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
CONTRACT_DIR=$PROJECT_DIR

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

# Clear .foc_env file
if [ -f .foc_env ]; then
  rm -f .foc_env
fi
touch .foc_env

echo "Deploying App contracts to devnet"
for entry in $(echo $APPS | jq -r '. | @base64'); do
  _jq() {
    echo ${entry} | base64 --decode | jq -r ${1}
  }
  APP_NAME=$(_jq '.name')
  APP_NAME_UPPER_UNDERSCORE=$(echo $APP_NAME | tr '[:lower:]' '[:upper:]' | tr '-' '_' | tr ' ' '_')
  APP_CLASS_NAME=$(_jq '.class_name')
  APP_CONSTRUCTOR_ARGS_RAW=$(_jq '.constructor_args[]' | tr '\n' ' ')
  if [ -z "$APP_CONSTRUCTOR_ARGS_RAW" ]; then
    APP_CONSTRUCTOR_ARGS=""
  else
    APP_CONSTRUCTOR_ARGS=$(echo $APP_CONSTRUCTOR_ARGS_RAW | sed "s/\$ACCOUNT/$DEVNET_ACCOUNT_ADDRESS/g")
    APP_CONSTRUCTOR_ARGS=$(echo --constructor-calldata "$APP_CONSTRUCTOR_ARGS")
  fi
  # TODO: Other replacements

  echo "Deploying contract \"$APP_CLASS_NAME\" to devnet"
  echo "cd $CONTRACT_DIR && sncast --accounts-file $DEVNET_ACCOUNT_FILE --account $DEVNET_ACCOUNT_NAME --wait --json declare --contract-name $APP_CLASS_NAME --url $RPC_URL"
  APP_DECLARE_RESULT=$(cd $CONTRACT_DIR && sncast --accounts-file $DEVNET_ACCOUNT_FILE --account $DEVNET_ACCOUNT_NAME --wait --json declare --contract-name $APP_CLASS_NAME --url $RPC_URL | tail -n 1)
  APP_CLASS_HASH=$(echo $APP_DECLARE_RESULT | jq -r '.class_hash')
  echo "Declared contract class hash: $APP_CLASS_HASH"

  echo "Deploying contract \"$APP_CLASS_NAME\" to devnet"
  echo "cd $CONTRACT_DIR && sncast --accounts-file $DEVNET_ACCOUNT_FILE --account $DEVNET_ACCOUNT_NAME --wait --json deploy --contract-name $APP_CLASS_NAME --url $RPC_URL --class-hash $APP_CLASS_HASH $APP_CONSTRUCTOR_ARGS"
  APP_DEPLOY_RESULT=$(cd $CONTRACT_DIR && sncast --accounts-file $DEVNET_ACCOUNT_FILE --account $DEVNET_ACCOUNT_NAME --wait --json deploy --url $RPC_URL --class-hash $APP_CLASS_HASH $APP_CONSTRUCTOR_ARGS | tail -n 1)
  APP_CONTRACT_ADDRESS=$(echo $APP_DEPLOY_RESULT | jq -r '.contract_address')
  echo "Deployed contract address: $APP_CONTRACT_ADDRESS"
  # TODO: Change where this is stored
  echo "${APP_NAME_UPPER_UNDERSCORE}_CONTRACT_ADDRESS=$APP_CONTRACT_ADDRESS" >> .foc_env

  # TODO: Use other methods exposed by the registry?
  echo "Registering contract \"$APP_CLASS_NAME\" to foc registry"
  echo "cd $CONTRACT_DIR && sncast --accounts-file $DEVNET_ACCOUNT_FILE --account $DEVNET_ACCOUNT_NAME --wait --json invoke --url $RPC_URL --contract-address $FOC_REGISTRY_CONTRACT_ADDRESS --function register_contract --calldata $APP_CONTRACT_ADDRESS $APP_CLASS_HASH"
  cd $CONTRACT_DIR && sncast --accounts-file $DEVNET_ACCOUNT_FILE --account $DEVNET_ACCOUNT_NAME --wait --json invoke --url $RPC_URL --contract-address $FOC_REGISTRY_CONTRACT_ADDRESS --function register_contract --calldata $APP_CONTRACT_ADDRESS $APP_CLASS_HASH
done

# TODO: Provide starkli option ?
