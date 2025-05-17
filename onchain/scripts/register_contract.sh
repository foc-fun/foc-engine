#!/bin/bash
#
# Register class with FocRegistry deployed in docker devnet

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
CONTRACT_DIR=$SCRIPT_DIR/..

ETH_ADDRESS=0x49D36570D4E46F48E99674BD3FCC84644DDD6B96F7C741B1562B82F9E004DC7
DEVNET_ACCOUNT_ADDRESS=0x064b48806902a367c8598f4f95c305e8c1a1acba5f082d294a43793113115691
DEVNET_ACCOUNT_NAME="account-1"
DEVNET_ACCOUNT_FILE=$CONTRACT_DIR/oz_acct.json

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

FOC_REGISTRY_CONTRACT=$1
CONTRACT_ADDRESS=$2
CLASS_HASH=$3
if [ -z "$FOC_REGISTRY_CONTRACT" ]; then
  echo "Usage: $0 <FOC_REGISTRY_CONTRACT> <CONTRACT_ADDRESS> <CLASS_HASH>"
  exit 1
fi
if [ -z "$CONTRACT_ADDRESS" ]; then
  echo "Usage: $0 <FOC_REGISTRY_CONTRACT> <CONTRACT_ADDRESS> <CLASS_HASH>"
  exit 1
fi
if [ -z "$CLASS_HASH" ]; then
  echo "Usage: $0 <FOC_REGISTRY_CONTRACT> <CONTRACT_ADDRESS> <CLASS_HASH>"
  exit 1
fi
CALLDATA=$(echo -n $CONTRACT_ADDRESS $CLASS_HASH)

echo "sncast --accounts-file $DEVNET_ACCOUNT_FILE --account $DEVNET_ACCOUNT_NAME --wait --json invoke --contract-address $FOC_REGISTRY_CONTRACT --function register_contract --calldata $CALLDATA --url $RPC_URL"
sncast --accounts-file $DEVNET_ACCOUNT_FILE --account $DEVNET_ACCOUNT_NAME --wait --json invoke --contract-address $FOC_REGISTRY_CONTRACT --function register_contract --calldata $CALLDATA --url $RPC_URL

# TODO: Provide starkli option ?
