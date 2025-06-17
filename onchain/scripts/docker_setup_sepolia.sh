#!/bin/bash
#
# Setup sepolia backend

# TODO: Properly check if devnet is running
# Wait for devnet to be up
sleep 15

ENGINE_HOST="indexer"
ENGINE_PORT=8085
ENGINE_URL=http://$ENGINE_HOST:$ENGINE_PORT
ENGINE_API_HOST="api"
ENGINE_API_PORT=8080
ENGINE_API_URL=http://$ENGINE_API_HOST:$ENGINE_API_PORT

FOC_ENGINE_CONFIG="/configs/sepolia.config.yaml"
FOC_REGISTRY_CONTRACT_ADDRESS=$(cat $FOC_ENGINE_CONFIG | yq -r '.Onchain.RegistryAddress')
FOC_ACCOUNTS_CONTRACT_ADDRESS=$(cat $FOC_ENGINE_CONFIG | yq -r '.Onchain.AccountsAddress')
FOC_ACCOUNTS_CLASS_HASH=$(cat $FOC_ENGINE_CONFIG | yq -r '.Onchain.AccountsClassHash')

echo "Setting up Foc Registry contract address: $FOC_REGISTRY_CONTRACT_ADDRESS"
echo "curl -X POST $ENGINE_URL/registry/add-registry-contract -d {\"address\":\"$FOC_REGISTRY_CONTRACT_ADDRESS\"}"
curl -X POST $ENGINE_URL/registry/add-registry-contract -d "{\"address\":\"$FOC_REGISTRY_CONTRACT_ADDRESS\",\"subscribeEvents\":\"true\"}"
curl -X POST $ENGINE_API_URL/registry/add-registry-contract -d "{\"address\":\"$FOC_REGISTRY_CONTRACT_ADDRESS\"}"

echo "Setting up Foc Accounts contract address: $FOC_ACCOUNTS_CONTRACT_ADDRESS"
echo "curl -X POST $ENGINE_URL/accounts/add-accounts-contract -d {\"address\":\"$FOC_ACCOUNTS_CONTRACT_ADDRESS\",\"class_hash\":\"$FOC_ACCOUNTS_CLASS_HASH\"}"
curl -X POST $ENGINE_URL/accounts/add-accounts-contract -d "{\"address\":\"$FOC_ACCOUNTS_CONTRACT_ADDRESS\",\"class_hash\":\"$FOC_ACCOUNTS_CLASS_HASH\",\"subscribeEvents\":\"true\"}"
curl -X POST $ENGINE_API_URL/accounts/add-accounts-contract -d "{\"address\":\"$FOC_ACCOUNTS_CONTRACT_ADDRESS\",\"class_hash\":\"$FOC_ACCOUNTS_CLASS_HASH\"}"
