#!/bin/bash
#
# Deploy & setup devnet contracts

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
PROJECT_DIR=$SCRIPT_DIR/../..
CONTRACT_DIR=$PROJECT_DIR/onchain

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

FOC_REGISTRY_CLASS_NAME="FocRegistry"
VERSION="0.0.1"
VERSION_UTF8_HEX=$(echo -n $VERSION | xxd -p -c 1000)
CALLDATA=$(echo -n 0x$VERSION_UTF8_HEX)

# TODO: Issue if already declared
echo "Deploying contract \"$FOC_REGISTRY_CLASS_NAME\" to devnet"
# echo "cd $CONTRACT_DIR && sncast --accounts-file $DEVNET_ACCOUNT_FILE --account $DEVNET_ACCOUNT_NAME --wait --json declare --contract-name $FOC_REGISTRY_CLASS_NAME --url $RPC_URL"
FOC_REGISTRY_CLASS_DECLARE_RESULT=$(cd $CONTRACT_DIR && sncast --accounts-file $DEVNET_ACCOUNT_FILE --account $DEVNET_ACCOUNT_NAME --wait --json declare --contract-name $FOC_REGISTRY_CLASS_NAME --url $RPC_URL | tail -n 1)
FOC_REGISTRY_CLASS_HASH=$(echo $FOC_REGISTRY_CLASS_DECLARE_RESULT | jq -r '.class_hash')
echo "Declared contract class hash: $FOC_REGISTRY_CLASS_HASH"

# echo "cd $CONTRACT_DIR && sncast --accounts-file $DEVNET_ACCOUNT_FILE --account $DEVNET_ACCOUNT_NAME --wait --json deploy --contract-name $FOC_REGISTRY_CLASS_NAME --url $RPC_URL --class-hash $FOC_REGISTRY_CLASS_HASH --constructor-args $CALLDATA"
FOC_REGISTRY_DEPLOY_RESULT=$(cd $CONTRACT_DIR && sncast --accounts-file $DEVNET_ACCOUNT_FILE --account $DEVNET_ACCOUNT_NAME --wait --json deploy --url $RPC_URL --class-hash $FOC_REGISTRY_CLASS_HASH --constructor-calldata $CALLDATA | tail -n 1)
FOC_REGISTRY_CONTRACT_ADDRESS=$(echo $FOC_REGISTRY_DEPLOY_RESULT | jq -r '.contract_address')
echo "Deployed contract address: $FOC_REGISTRY_CONTRACT_ADDRESS"

# TODO: Provide starkli option ?
# echo "starkli declare --rpc $RPC_URL --account $DEVNET_ACCOUNT_FILE --private-key $DEVNET_ACCOUNT_PRIVATE_KEY --casm-file $POW_CONTRACT_CASM_FILE $POW_CONTRACT_SIERRA_FILE"
# POW_DECLARE_OUTPUT=$(starkli declare --rpc $RPC_URL --account $DEVNET_ACCOUNT_FILE --private-key $DEVNET_ACCOUNT_PRIVATE_KEY --casm-file $POW_CONTRACT_CASM_FILE $POW_CONTRACT_SIERRA_FILE > $LOG_DIR/deploy.log 2>&1)
# POW_CONTRACT_CLASSHASH=$(echo $POW_DECLARE_OUTPUT | tail -n 1 | awk '{print $NF}')
# echo "Contract class hash: $POW_CONTRACT_CLASSHASH"
# 
# echo "Deploying contract \"$FOC_REGISTRY_CLASS_NAME\" to devnet"
# echo "starkli deploy --rpc $RPC_URL --private-key $DEVNET_ACCOUNT_PRIVATE_KEY $POW_CONTRACT_CLASSHASH $CALLDATA"
# POW_DEPLOY_OUTPUT=$(starkli deploy --rpc $RPC_URL --private-key $DEVNET_ACCOUNT_PRIVATE_KEY $POW_CONTRACT_CLASSHASH $CALLDATA > $LOG_DIR/deploy.log 2>&1)
# FOC_REGISTRY_CONTRACT_ADDRESS=$(echo $POW_DEPLOY_OUTPUT | tail -n 1 | awk '{print $NF}')
# echo "Contract address: $FOC_REGISTRY_CONTRACT_ADDRESS"
