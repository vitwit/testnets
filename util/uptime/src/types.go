package src

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
}