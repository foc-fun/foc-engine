use starknet::{ContractAddress};

#[derive(Drop, Serde)]
pub struct FocAccount {
  pub contract: ContractAddress,
  pub user: ContractAddress,
  pub username: felt252,
  pub metadata: Span<felt252>,
}

#[starknet::interface]
pub trait IFocAccounts<TContractState> {
  fn get_username(self: @TContractState, user: ContractAddress) -> felt252;
  fn get_contract_username(self: @TContractState, contract: ContractAddress, user: ContractAddress) -> felt252;
  fn get_account(self: @TContractState, user: ContractAddress) -> FocAccount;
  fn get_contract_account(self: @TContractState, contract: ContractAddress, user: ContractAddress) -> FocAccount;
  fn get_accounts(self: @TContractState, users: Span<ContractAddress>) -> Span<FocAccount>;
  fn get_contract_accounts(self: @TContractState, contract: ContractAddress, users: Span<ContractAddress>) -> Span<FocAccount>;
  fn is_username_claimed(self: @TContractState, username: felt252) -> bool;
  fn is_contract_username_claimed(self: @TContractState, contract: ContractAddress, username: felt252) -> bool;

  fn claim_username(ref self: TContractState, username: felt252);
  fn claim_contract_username(ref self: TContractState, contract: ContractAddress, username: felt252);
  fn set_account_metadata(ref self: TContractState, account_metadata: Span<felt252>);
  fn set_contract_account_metadata(ref self: TContractState, contract: ContractAddress, account_metadata: Span<felt252>);
}

#[starknet::contract]
pub mod FocAccounts {
  use starknet::{ContractAddress, get_caller_address, get_contract_address};
  use starknet::storage::{Map, StorageMapReadAccess, StorageMapWriteAccess, StoragePathEntry, StoragePointerReadAccess, StoragePointerWriteAccess, Vec, MutableVecTrait, VecTrait};
  use super::{FocAccount, IFocAccounts};

  // TODO: Implement these at the component level and allow people to customize for their own contract within that contract before claim_contract_username is called
  // #[derive(Drop, Serde, starknet::Store, Clone)ed
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

  #[derive(Drop, starknet::Event)]
  struct AccountMetadataSet {
    #[key]
    contract: ContractAddress,
    #[key]
    user: ContractAddress,
    metadata: Span<felt252>,
  }

  #[event]
  #[derive(Drop, starknet::Event)]
  pub enum Event {
    UsernameClaimed: UsernameClaimed,
    AccountMetadataSet: AccountMetadataSet,
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

    fn get_account(self: @ContractState, user: ContractAddress) -> FocAccount {
      let contract = get_contract_address();
      let account_metadata_vec = self.contract_accounts_metadata.entry((contract, user));
      let mut account_metadata = array![];
      let mut idx: u32 = 0;
      let account_metadata_len = account_metadata_vec.len().try_into().unwrap();
      while idx != account_metadata_len {
        let storage_ptr = account_metadata_vec.at(idx.into());
        account_metadata.append(storage_ptr.read());
        idx += 1;
      }
      return FocAccount {
        contract: contract,
        user: user,
        username: self.get_username(user),
        metadata: account_metadata.span()
      };
    }

    fn get_contract_account(self: @ContractState, contract: ContractAddress, user: ContractAddress) -> FocAccount {
      let account_metadata_vec = self.contract_accounts_metadata.entry((contract, user));
      let mut account_metadata = array![];
      let mut idx: u32 = 0;
      let account_metadata_len = account_metadata_vec.len().try_into().unwrap();
      while idx != account_metadata_len {
        let storage_ptr = account_metadata_vec.at(idx.into());
        account_metadata.append(storage_ptr.read());
        idx += 1;
      }
      return FocAccount {
        contract: contract,
        user: user,
        username: self.get_contract_username(contract, user),
        metadata: account_metadata.span(),
      };
    }

    fn get_accounts(self: @ContractState, users: Span<ContractAddress>) -> Span<FocAccount> {
      let mut accounts = array![];
      for user in users {
        let account = self.get_account(*user);
        accounts.append(account);
      }
      return accounts.into();
    }

    fn get_contract_accounts(self: @ContractState, contract: ContractAddress, users: Span<ContractAddress>) -> Span<FocAccount> {
      let mut accounts = array![];
      for user in users {
        let account = self.get_contract_account(contract, *user);
        accounts.append(account);
      }
      return accounts.into();
    }

    fn is_username_claimed(self: @ContractState, username: felt252) -> bool {
      let contract = get_contract_address();
      let contract_username_owners = self.contract_username_owners.read((contract, username));
      return contract_username_owners.into() != 0;
    }

    fn is_contract_username_claimed(self: @ContractState, contract: ContractAddress, username: felt252) -> bool {
      let contract_username_owners = self.contract_username_owners.read((contract, username));
      return contract_username_owners.into() != 0;
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

    fn claim_contract_username(ref self: ContractState, contract: ContractAddress, username: felt252) {
      // This function is only callable from the contract
      let user = get_caller_address();

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

    fn set_account_metadata(ref self: ContractState, account_metadata: Span<felt252>) {
      let contract = get_contract_address();
      let user = get_caller_address();
      // TODO: Clear users account metadata if it exists already?
      let mut idx: u32 = 0;
      let account_metadata_len = account_metadata.len().try_into().unwrap();
      while idx != account_metadata_len {
        let metadata_ptr = self.contract_accounts_metadata.entry((contract, user));
        let storage_vec_len: u64 = metadata_ptr.len();
        if idx.into() >= storage_vec_len {
          metadata_ptr.push(*account_metadata.at(idx));
        } else {
          let mut storage_ptr = metadata_ptr.at(idx.into());
          storage_ptr.write(*account_metadata.at(idx));
        }
        idx += 1;
      }
    }

    fn set_contract_account_metadata(ref self: ContractState, contract: ContractAddress, account_metadata: Span<felt252>) {
      let user = get_caller_address();
      // TODO: Clear users account metadata if it exists already?
      let mut idx = 0;
      while idx != account_metadata.len() {
        let metadata_ptr = self.contract_accounts_metadata.entry((contract, user));
        let storage_vec_len: u64 = metadata_ptr.len();
        if idx.into() >= storage_vec_len {
          metadata_ptr.push(*account_metadata.at(idx));
        } else {
          let mut storage_ptr = metadata_ptr.at(idx.into());
          storage_ptr.write(*account_metadata.at(idx));
        }
        idx += 1;
      }
    }
  }
}
