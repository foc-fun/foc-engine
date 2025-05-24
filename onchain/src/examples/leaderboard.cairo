use starknet::{ContractAddress};

#[starknet::interface]
pub trait ILeaderboardContract<TContractState> {
    fn get_user_score(self: @TContractState, user: ContractAddress) -> felt252;
    fn set_user_score(ref self: TContractState, user: ContractAddress, score: felt252);
}

#[starknet::contract]
pub mod LeaderboardContract {
    use starknet::{ContractAddress};
    use starknet::storage::{Map, StorageMapReadAccess, StorageMapWriteAccess};

    #[storage]
    struct Storage {
        // Maps: user address -> score
        user_scores: Map<ContractAddress, felt252>,
    }

    #[derive(Drop, Serde, starknet::Event)]
    pub struct ScoreChanged {
        #[key]
        user: ContractAddress,
        old_score: felt252,
        new_score: felt252,
    }

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {
        ScoreChanged: ScoreChanged,
    }

    #[constructor]
    fn constructor(ref self: ContractState) {
    }

    #[abi(embed_v0)]
    impl LeaderboardContractImpl of super::ILeaderboardContract<ContractState> {
        fn get_user_score(self: @ContractState, user: ContractAddress) -> felt252 {
            self.user_scores.read(user)
        }

        fn set_user_score(ref self: ContractState, user: ContractAddress, score: felt252) {
            let old_score = self.user_scores.read(user);
            self.user_scores.write(user, score);
            self.emit(ScoreChanged { 
                user: user,
                old_score: old_score,
                new_score: score,
            });
        }
    }
}
