package src

import "time"

type Validator struct {
	ValidatorInfo []ValidatorInfo `json:"validatorInfo"`
}

type ValidatorInfo struct {
	ValAddress string `json:"valAddress"`
	Info       Info   `json:"info"`
}

type Info struct {
	UptimePoints       float64 `json:"uptimePoints"`
	Moniker            string  `json:"moniker"`
	OperatorAddr       string  `json:"operatorAddr"`
	Upgrade1Points     int64   `json:"upgrade1Points"`
	Upgrade2Points     int64   `json:"upgrade2Points"`
	Upgrade3Points     int64   `json:"upgrade3Points"`
	Upgrade4Points     int64   `json:"upgrade4Points"`
	StartBlock         int64   `json:"startBlock"`
	UptimeCount        int64   `json:"uptimeCount"`
	GenesisPoints      int64   `json:"genesisPoints"`
	TotalPoints        float64 `json:"totalPoints"`
	Proposal1VoteScore int64   `json:"proposal1VoteScore"`
	Proposal2VoteScore int64   `json:"proposal2VoteScore"`
	Proposal3VoteScore int64   `json:"proposal3VoteScore"`
	Proposal4VoteScore int64   `json:"proposal4VoteScore"`
	DelegatorAddress   string  `json:"delegator_address"`
}

type Proposal struct {
	ProposalContent struct {
		Type  string `json:"type"`
		Value struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Plan        struct {
				Name string    `json:"name"`
				Time time.Time `json:"time"`
			} `json:"plan"`
		} `json:"value"`
	} `json:"content"`
	ProposalID       string `json:"id"`
	ProposalStatus   string `json:"proposal_status"`
	FinalTallyResult struct {
		Yes        string `json:"yes"`
		Abstain    string `json:"abstain"`
		No         string `json:"no"`
		NoWithVeto string `json:"no_with_veto"`
	} `json:"final_tally_result"`
	SubmitTime     time.Time `json:"submit_time"`
	DepositEndTime time.Time `json:"deposit_end_time"`
	TotalDeposit   struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	} `json:"total_deposit"`
	VotingStartTime time.Time `json:"voting_start_time"`
	VotingEndTime   time.Time `json:"voting_end_time"`
}