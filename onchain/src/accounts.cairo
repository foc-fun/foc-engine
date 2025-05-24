use starknet::{ContractAddress};

#[starknet::interface]
pub trait IFocAccounts<TContractState> {
  fn get_username(self: @TContractState, user: ContractAddress) -> felt252;
  fn get_contract_username(self: @TContractState, contract: ContractAddress, user: ContractAddress) -> felt252;

  fn claim_username(ref self: TContractState, username: felt252);
  fn claim_contract_username(ref self: TContractState, user: ContractAddress, username: felt252);
}

#[starknet::contract]
pub mod FocAccounts {
  use starknet::{ContractAddress, get_caller_address, get_contract_address};
  use starknet::storage::{Map, StorageMapReadAccess, StorageMapWriteAccess, StoragePointerReadAccess, StoragePointerWriteAccess, Vec};

  // TODO: Implement these at the component level and allow people to customize for their own contract within that contract before claim_contract_username is called
  // #[derive(Drop, Serde, starknet::Store, Clone)]
  // struct ContractMetadataConfig {
  //   // Username configs
  //   max_username_length: Option<felt252>,
  //   disallowed_usernames: Option<Span<felt252>>,
  //   disallowed_characters: Option<Span<felt252>>,
  //   // Account metadata configs
  //   max_account_metadata_length: Option<felt252>,
  //   // TODO: Add more configurable restrictions like Max/min per metadata value, uniqueness constraints, hardcode/functional based constraints, ...
  // }

  #[derive(Drop, Serde)]
  pub struct InitParams {
      version: felt252,
  }

  #[storage]
  struct Storage {
    version: felt252,
    // Mapping: (contract, user) -> username
    contract_usernames: Map<(ContractAddress, ContractAddress), felt252>,
    // Mapping: (contract, username) -> user
    contract_username_owners: Map<(ContractAddress, felt252), ContractAddress>,

    // Mapping: (contract, user) -> account metadata
    contract_accounts_metadata: Map<(ContractAddress, ContractAddress), Vec<felt252>>,
    // Mapping: contract -> metadata configuration
    // contract_metadata_config: Map<ContractAddress, ContractMetadataConfig>,
  }

  #[derive(Drop, starknet::Event)]
  struct UsernameClaimed {
    #[key]
    contract: ContractAddress,
    #[key]
    user: ContractAddress,
    username: felt252,
  }

  #[event]
  #[derive(Drop, starknet::Event)]
  pub enum Event {
    UsernameClaimed: UsernameClaimed,
  }

  #[constructor]
  fn constructor(ref self: ContractState, init_params: InitParams) {
      self.version.write(init_params.version);
  }

  #[abi(embed_v0)]
  impl FocAccountsImpl of super::IFocAccounts<ContractState> {
    fn get_username(self: @ContractState, user: ContractAddress) -> felt252 {
      let username = self.contract_usernames.read((get_contract_address(), user));
      return username;
    }

    fn get_contract_username(self: @ContractState, contract: ContractAddress, user: ContractAddress) -> felt252 {
      let username = self.contract_usernames.read((contract, user));
      return username;
    }

    fn claim_username(ref self: ContractState, username: felt252) {
      let contract = get_contract_address();
      let user = get_caller_address();

      // Check if the user already has a username
      // TODO: Clear users username data if it exists already instead?
      let contract_usernames = self.contract_usernames.read((contract, user));
      assert!(contract_usernames == 0, "Username already claimed");
      // Check if the username is already taken
      let contract_username_owners = self.contract_username_owners.read((contract, username));
      assert!(contract_username_owners.into() == 0, "Username already taken");
      // Check if the username is valid
      // TODO: Implement username validation logic from contract username config ( ex: no :: )
      // TODO: Implement username validation logic from metadata config

      // Store the username
      self.contract_usernames.write((contract, user), username);
      self.contract_username_owners.write((contract, username), user);
      self.emit(UsernameClaimed {
        contract: contract,
        user: user,
        username: username,
      });
    }

    fn claim_contract_username(ref self: ContractState, user: ContractAddress, username: felt252) {
      // This function is only callable from the contract
      let contract = get_caller_address();

      // Check if the user already has a username
      // TODO: Clear users username data if it exists already instead?
      let contract_usernames = self.contract_usernames.read((contract, user));
      assert!(contract_usernames == 0, "Username already claimed");
      // Check if the username is already taken
      let contract_username_owners = self.contract_username_owners.read((contract, username));
      assert!(contract_username_owners.into() == 0, "Username already taken");
      // Check if the username is valid
      // TODO

      // Store the username
      self.contract_usernames.write((contract, user), username);
      self.contract_username_owners.write((contract, username), user);
      self.emit(UsernameClaimed {
        contract: contract,
        user: user,
        username: username,
      });
    }
  }
}
