package db

import (
	"github.com/spf13/viper"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

//configuring db name and collections
var (
	DB_NAME, dbErr        = viper.Get("database").(string)
	BLOCKS_COLLECTION     = "blocks"
	VALIDATORS_COLLECTION = "validators"
	PROPOSALS_COLLECTION  = "proposals"
)

type Blocks struct {
	ID         string   `json:"_id" bson:"_id"`
	Height     int64    `json:"height" bson:"height"`
	Validators []string `json:"validators" bson:"validators"`
}

type Validator struct {
	Address         string      `json:"address" bson:"address"`
	OperatorAddress string      `json:"operatorAddress" bson:"operator_address"`
	Description     Description `json:"description" bson:"description"`
}

type ValAggregateResult struct {
	Id                string              `json:"_id" bson.M:"_id"`
	Uptime_count      int64               `json:"uptime_count" bson:"uptime_count"`
	Upgrade1_block    int64               `json:"upgrade1_block" bson:"upgrade1_block"`
	Upgrade2_block    int64               `json:"upgrade2_block" bson:"upgrade2_block"`
	Upgrade3_block    int64               `json:"upgrade3_block" bson:"upgrade3_block"`
	Upgrade4_block    int64               `json:"upgrade4_block" bson:"upgrade4_block"`
	Validator_details []Validator_details `json:"validator_details" bson:"validator_details"`
}

type Validator_details struct {
	Description       Description `json:"description" bson:"description"`
	Operator_address  string      `json:"operator_address" bson:"operator_address"`
	Address           string      `json:"address" bson:"address"`
	Delegator_address string      `json:"delegator_address" bson:"delegator_address"`
}

type Description struct {
	Moniker string `json:"moniker" bson:"moniker"`
}

//type Proposals struct {
//	Voter       string `json:"voter" bson:"voter"`
//	Proposal_id string `json:"proposal_id" bson:"proposal_id"`
//	Option      string `json:"option" bson:"option"`
//}

type Proposals struct {
	ProposalID      int       `json:"proposalId" bson:"proposalId"`
	ID              string    `json:"id" bson:"id"`
	ProposalStatus  string    `json:"proposal_status" bson:"proposal_status"`
	Proposer        string    `json:"proposer" bson:"proposer"`
	SubmitTime      time.Time `json:"submit_time" bson:"submit_time"`
	VotingEndTime   time.Time `json:"voting_end_time" bson:"voting_end_time"`
	VotingStartTime time.Time `json:"voting_start_time" bson:"voting_start_time"`
	UpdatedAt       string    `json:"updatedAt" bson:"updatedAt"`
	Votes           []struct {
		Voter       string `json:"voter" bson:"voter"`
		ProposalID  string `json:"proposal_id" bson:"proposal_id"`
		Option      string `json:"option" bson:"option"`
		VotingPower int    `json:"votingPower" bson:"votingPower"`
	} `json:"votes" bson:"votes"`
	Value struct {
		ProposalStatus string `json:"proposal_status" bson:"proposal_status"`
	} `json:"value" bson:"value"`
}

// Connect returns a pointer to a MongoDB instance,
// which is used for collecting the metrics required for uptime calculations
func Connect(info *mgo.DialInfo) (DB, error) {
	session, err := mgo.DialWithInfo(info)

	return Store{session: session}, err
}

// Terminate should be used to terminate a database session, generally in a defer statement inside main app file.
func (db Store) Terminate() {
	db.session.Close()
}

//QueryValAggregateData - Fetch all blocks by using aggregate query
func (db Store) QueryValAggregateData(aggQuery []bson.M) (result []ValAggregateResult, err error) {
	err = db.session.DB(DB_NAME).C(BLOCKS_COLLECTION).Pipe(aggQuery).All(&result)
	return result, err
}

func (db Store) QueryProposalDetails(query bson.M) (proposals Proposals, err error) {
	err = db.session.DB(DB_NAME).C(PROPOSALS_COLLECTION).Find(query).One(&proposals)
	return proposals, err
}

type (
	// DB interface defines all the methods accessible by the application
	DB interface {
		Terminate()
		QueryValAggregateData(aggQuery []bson.M) ([]ValAggregateResult, error)
		QueryProposalDetails(query bson.M) (Proposals, error)
	}

	// Store will be used to satisfy the DB interface
	Store struct {
		session *mgo.Session
	}
)
