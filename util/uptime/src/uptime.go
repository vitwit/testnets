package src

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"text/tabwriter"

	"github.com/kataras/golog"
	"github.com/spf13/viper"
	"github.com/vitwit/testnets/util/uptime/db"
	"gopkg.in/mgo.v2/bson"
)

var (
	papuaStartBlock     int64
	papuaEndBlock       int64
	papuaPointsPerBlock int64

	patagoniaStartBlock     int64
	patagoniaEndBlock       int64
	patagoniaPointsPerBlock int64

	dariengapStartBlock     int64
	dariengapEndBlock       int64
	dariengapPointsPerBlock int64

	andesStartBlock     int64
	andesEndBlock       int64
	andesPointsPerBlock int64

	nodeRewards int64
)

type handler struct {
	db db.DB
}

func New(db db.DB) handler {
	return handler{db}
}

func (h *handler) CalculateProposalsVoteScore(proposal_id string, delegator_address string) int64 {
	query := bson.M{
		"proposal_id": proposal_id,
		"voter":       delegator_address,
	}

	proposal, _ := h.db.QueryProposalDetails(query)

	if proposal.Voter != "" {
		return 50
	}

	return 0
}

func GenerateAggregateQuery(startBlock int64, endBlock int64,
	patagoniaStartBlock int64, patagoniaEndBlock int64,
	papuaStartBlock int64, papuaEndBlock int64,
	dariengapStartBlock int64, dariengapEndBlock int64,
	andesStartBlock int64, andesEndBlock int64) []bson.M {
	aggQuery := []bson.M{}

	//Query for filtering blocks in between given start block and end block
	matchQuery := bson.M{
		"$match": bson.M{
			"$and": []bson.M{
				bson.M{
					"height": bson.M{"$gte": startBlock},
				},
				bson.M{
					"height": bson.M{"$lte": endBlock},
				},
			},
		},
	}

	aggQuery = append(aggQuery, matchQuery)

	//Query for Unwind the Array of validators from each block
	unwindQuery := bson.M{
		"$unwind": "$validators",
	}

	aggQuery = append(aggQuery, unwindQuery)

	//Query for calculating uptime count, upgrade1, upgrade2, upgrade3 and upgrade4 count
	groupQuery := bson.M{
		"$group": bson.M{
			"_id":          "$validators",
			"uptime_count": bson.M{"$sum": 1},
			"upgrade1_block": bson.M{
				"$min": bson.M{
					"$cond": []interface{}{
						bson.M{
							"$and": []bson.M{
								bson.M{"$gte": []interface{}{"$height", patagoniaStartBlock}},
								bson.M{"$lte": []interface{}{"$height", patagoniaEndBlock}},
							},
						},
						"$height",
						"null",
					},
				},
			},
			"upgrade2_block": bson.M{
				"$min": bson.M{
					"$cond": []interface{}{
						bson.M{
							"$and": []bson.M{
								bson.M{"$gte": []interface{}{"$height", papuaStartBlock}},
								bson.M{"$lte": []interface{}{"$height", papuaEndBlock}},
							},
						},
						"$height",
						"null",
					},
				},
			},
			"upgrade3_block": bson.M{
				"$min": bson.M{
					"$cond": []interface{}{
						bson.M{
							"$and": []bson.M{
								bson.M{"$gte": []interface{}{"$height", dariengapStartBlock}},
								bson.M{"$lte": []interface{}{"$height", dariengapEndBlock}},
							},
						},
						"$height",
						"null",
					},
				},
			},
			"upgrade4_block": bson.M{
				"$min": bson.M{
					"$cond": []interface{}{
						bson.M{
							"$and": []bson.M{
								bson.M{"$gte": []interface{}{"$height", andesStartBlock}},
								bson.M{"$lte": []interface{}{"$height", andesEndBlock}},
							},
						},
						"$height",
						"null",
					},
				},
			},
		},
	}

	aggQuery = append(aggQuery, groupQuery)

	//Query for getting moniker, operator address from validators
	lookUpQuery := bson.M{
		"$lookup": bson.M{
			"from": "validators",
			"let":  bson.M{"id": "$_id"},
			"pipeline": []bson.M{
				bson.M{
					"$match": bson.M{
						"$expr": bson.M{"$eq": []string{"$address", "$$id"}},
					},
				},
				bson.M{
					"$project": bson.M{
						"description.moniker": 1, "operator_address": 1, "delegator_address": 1, "address": 1, "_id": 0,
					},
				},
			},
			"as": "validator_details",
		},
	}

	aggQuery = append(aggQuery, lookUpQuery)

	res2B, _ := json.Marshal(aggQuery)
	fmt.Println(string(res2B))

	return aggQuery
}

// CalculateUpgradePoints - Calculates upgrade points by using upgrade points per block,
// upgrade block and end block height
func CalculateUpgradePoints(startBlock int64, valUpgradeBlock int64, endBlockHeight int64, totalScore int64, missedDeductionFactor int64) int64 {
	var rewards int64 = 0

	if valUpgradeBlock == 0 {
		fmt.Println("valUpgradeBlock-----": valUpgradeBlock, missedDeductionFactor)
		return rewards
	}

	if valUpgradeBlock == startBlock {
		rewards = totalScore
	} else if (endBlockHeight - valUpgradeBlock) > 0 {
		rewards = totalScore - ((valUpgradeBlock - startBlock) * missedDeductionFactor) // each missed block costs 1 point deduction
	}

	fmt.Println(missedDeductionFactor, valUpgradeBlock, rewards)

	return rewards
}

func GetCommonValidators(gentxVals, blockVals []string) (results []string) {
	m := make(map[string]bool)

	for _, item := range gentxVals {
		m[item] = true
	}

	for _, item := range blockVals {
		if _, ok := m[item]; ok {
			results = append(results, item)
		}
	}

	return results
}

func (h handler) CalculateGenesisPoints(address string) int64 {
	var aggQuery []bson.M

	matchQuery := bson.M{
		"$match": bson.M{
			"height": 2,
		},
	}

	aggQuery = append(aggQuery, matchQuery)

	lookUpQuery := bson.M{
		"$lookup": bson.M{
			"from": "validators",
			"let":  bson.M{"id": "$validators"},
			"pipeline": []bson.M{
				bson.M{
					"$match": bson.M{
						"$expr": bson.M{"$in": []string{"$address", "$$id"}},
					},
				},
				bson.M{
					"$project": bson.M{
						"description.moniker": 1, "operator_address": 1, "address": 1, "delegator_address": 1, "_id": 0,
					},
				},
			},
			"as": "validator_details",
		},
	}

	aggQuery = append(aggQuery, lookUpQuery)

	results, err := h.db.QueryValAggregateData(aggQuery)

	if err != nil {
		fmt.Printf("Error while fetching validator data at height 2 %v", err)
		db.HandleError(err)
	}

	var blockValidators []string

	if len(results) > 0 {
		for _, val := range results[0].Validator_details {
			blockValidators = append(blockValidators, val.Operator_address)
		}
	}

	for _, val := range blockValidators {
		if val == address {
			return 100
		}
	}

	return 0

}

// CalculateUptimeRewards uptime rewards max 200
func CalculateUptimeRewards(uptimeCount int64, startBlock int64, endBlock int64) float64 {
	totalBlocks := endBlock - startBlock

	if uptimeCount > totalBlocks {
		return -1
	}

	uptimePerc := float64(uptimeCount) / float64(totalBlocks) * 100

	if uptimePerc > 90 {
		return float64((uptimePerc-90)*10*200) / 100
	}

	return 0
}

func (h handler) CalculateUptime(startBlock int64, endBlock int64) {
	//Read node rewards from config

	//nodeRewards = viper.Get("node_rewards").(int64)

	// Read papua upgrade configs
	papuaStartBlock = viper.Get("papua_startblock").(int64) + 1 //Need to consider votes from next block after upgrade
	papuaEndBlock = viper.Get("papua_endblock").(int64) + 1
	//papuaPointsPerBlock = viper.Get("papua_reward_points_per_block").(int64)

	// Read patagonia upgrade configs
	patagoniaStartBlock = viper.Get("patagonia_startblock").(int64) + 1 //Need to consider votes from next block after upgrade
	patagoniaEndBlock = viper.Get("patagonia_endblock").(int64) + 1
	//patagoniaPointsPerBlock = viper.Get("patagonia_reward_points_per_block").(int64)

	// Read darien-gap upgrade configs
	dariengapStartBlock = viper.Get("darien_gap_startblock").(int64) + 1 //Need to consider votes from next block after upgrade
	dariengapEndBlock = viper.Get("darien_gap_endblock").(int64) + 1
	//dariengapPointsPerBlock = viper.Get("darien_gap_reward_points_per_block").(int64)

	// Read andes upgrade configs
	andesStartBlock = viper.Get("andes_startblock").(int64) + 1 //Need to consider votes from next block after upgrade
	andesEndBlock = viper.Get("andes_endblock").(int64) + 1
	//andesPointsPerBlock = viper.Get("patagonia_reward_points_per_block").(int64)

	var validatorsList []ValidatorInfo //Intializing validators uptime

	fmt.Println("Fetching blocks from:", startBlock, ", to:", endBlock)

	upgrade1AggQuery := GenerateAggregateQuery(startBlock, endBlock,
		patagoniaStartBlock, patagoniaEndBlock,
		papuaStartBlock, papuaEndBlock,
		dariengapStartBlock, dariengapEndBlock,
		andesStartBlock, andesEndBlock)

	results, err := h.db.QueryValAggregateData(upgrade1AggQuery)
	if err != nil {
		golog.Error("Error while fetching the data..", err)
	}

	if err != nil {
		fmt.Printf("Error while fetching validator data %v", err)
		db.HandleError(err)
	}

	for _, obj := range results {

		var specialBonus int64 = 0
		if obj.Validator_details[0].Operator_address == viper.Get("special_bonus_address").(string) {
			specialBonus = 100
		}

		valInfo := ValidatorInfo{
			ValAddress: obj.Validator_details[0].Address,
			Info: Info{
				OperatorAddr:     obj.Validator_details[0].Operator_address,
				Moniker:          obj.Validator_details[0].Description.Moniker,
				UptimeCount:      obj.Uptime_count,
				Upgrade1Points:   CalculateUpgradePoints(patagoniaStartBlock, obj.Upgrade1_block, patagoniaEndBlock, 150, 1),
				Upgrade2Points:   CalculateUpgradePoints(papuaStartBlock, obj.Upgrade2_block, papuaEndBlock, 150, 2),
				Upgrade3Points:   CalculateUpgradePoints(dariengapStartBlock, obj.Upgrade3_block, dariengapEndBlock, 150, 3) + specialBonus,
				Upgrade4Points:   CalculateUpgradePoints(andesStartBlock, obj.Upgrade4_block, andesEndBlock, 150, 5),
				UptimePoints:     CalculateUptimeRewards(obj.Uptime_count, startBlock, endBlock),
				DelegatorAddress: obj.Validator_details[0].Delegator_address,
			},
		}

		validatorsList = append(validatorsList, valInfo)
	}

	//calculating uptime points
	for i, v := range validatorsList {
		//maxUptimeRewards := viper.Get("max_uptime_rewards").(int64)
		//uptimePoints := float64(v.Info.UptimeCount*maxUptimeRewards) / (float64(endBlock) - float64(startBlock))

		validatorsList[i].Info.UptimePoints = v.Info.UptimePoints

		proposal1 := viper.Get("proposal_1_id").(string)
		proposal2 := viper.Get("proposal_2_id").(string)
		proposal3 := viper.Get("proposal_3_id").(string)
		proposal4 := viper.Get("proposal_4_id").(string)
		//calculate proposal1 vote score
		proposal1VoteScore := h.CalculateProposalsVoteScore(proposal1, validatorsList[i].Info.DelegatorAddress)

		//calculate proposal2 vote score
		proposal2VoteScore := h.CalculateProposalsVoteScore(proposal2, validatorsList[i].Info.DelegatorAddress)

		proposal3VoteScore := h.CalculateProposalsVoteScore(proposal3, validatorsList[i].Info.DelegatorAddress)

		proposal4VoteScore := h.CalculateProposalsVoteScore(proposal4, validatorsList[i].Info.DelegatorAddress)

		validatorsList[i].Info.Proposal1VoteScore = proposal1VoteScore
		validatorsList[i].Info.Proposal2VoteScore = proposal2VoteScore
		validatorsList[i].Info.Proposal3VoteScore = proposal3VoteScore
		validatorsList[i].Info.Proposal4VoteScore = proposal4VoteScore

		genesisPoints := h.CalculateGenesisPoints(validatorsList[i].Info.OperatorAddr)
		validatorsList[i].Info.GenesisPoints = genesisPoints

		validatorsList[i].Info.TotalPoints = float64(validatorsList[i].Info.Upgrade1Points) +
			float64(validatorsList[i].Info.Upgrade2Points) + float64(validatorsList[i].Info.Upgrade3Points) +
			float64(validatorsList[i].Info.Upgrade4Points) + v.Info.UptimePoints +
			float64(proposal1VoteScore) + float64(proposal2VoteScore) + float64(proposal3VoteScore) +
			float64(proposal4VoteScore) + float64(genesisPoints)

	}

	//Printing Uptime results in tabular view
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 0, ' ', tabwriter.Debug)
	fmt.Fprintln(w, " Operator Addr \t Moniker\t Uptime Count \t Uptime Points "+
		"\t Upgrade-1 Points \t Upgrade-2 Points \t Upgrade-3 Points \t Upgrade-4 Points "+
		" \t Proposal-1 Points \t Proposal-2 Points \t Proposal-3 Points \t Proposal-4 Points \t Genesis Points \t Total points")

	for _, data := range validatorsList {
		var address string = data.Info.OperatorAddr

		//Assigning validator address if operator address is not found
		if address == "" {
			address = data.ValAddress + " (Hex Address)"
		}

		fmt.Fprintln(w, " "+address+"\t "+data.Info.Moniker+
			"\t  "+strconv.Itoa(int(data.Info.UptimeCount))+" \t"+fmt.Sprintf("%f", data.Info.UptimePoints)+
			"\t "+strconv.Itoa(int(data.Info.Upgrade1Points))+" \t"+strconv.Itoa(int(data.Info.Upgrade2Points))+
			"\t "+strconv.Itoa(int(data.Info.Upgrade3Points))+" \t"+strconv.Itoa(int(data.Info.Upgrade4Points))+
			"\t"+strconv.Itoa(int(data.Info.Proposal1VoteScore))+"\t"+strconv.Itoa(int(data.Info.Proposal2VoteScore))+
			"\t"+strconv.Itoa(int(data.Info.Proposal3VoteScore))+"\t"+strconv.Itoa(int(data.Info.Proposal4VoteScore))+
			"\t"+strconv.Itoa(int(data.Info.GenesisPoints))+"\t"+fmt.Sprintf("%f", data.Info.TotalPoints))
	}

	w.Flush()

	//Export data to csv file
	ExportToCsv(validatorsList, nodeRewards)
}

// ExportToCsv - Export data to CSV file
func ExportToCsv(data []ValidatorInfo, nodeRewards int64) {
	Header := []string{
		"ValOper Address", "Moniker", "Uptime Count", "Uptime Points", "Upgrade1 Points",
		"Upgrade2 Points", "Upgrade3 Points", "Upgrade4 Points",
		"Proposal1 Vote Points", "Proposal2 Vote Points", "Proposal3 Vote Points", "Proposal4 Vote Points",
		"Genesis Points", "Total Points",
	}

	file, err := os.Create("result.csv")

	if err != nil {
		log.Fatal("Cannot write to file", err)
	}

	defer file.Close() //Close file

	writer := csv.NewWriter(file)

	defer writer.Flush()

	//Write header titles
	_ = writer.Write(Header)

	for _, record := range data {
		var address string = record.Info.OperatorAddr

		//Assigning validator address if operator address is not found
		if address == "" {
			address = record.ValAddress + " (Hex Address)"
		}

		uptimeCount := strconv.Itoa(int(record.Info.UptimeCount))
		uptimePoints := fmt.Sprintf("%f", record.Info.UptimePoints)
		up1Points := strconv.Itoa(int(record.Info.Upgrade1Points))
		up2Points := strconv.Itoa(int(record.Info.Upgrade2Points))
		up3Points := strconv.Itoa(int(record.Info.Upgrade3Points))
		up4Points := strconv.Itoa(int(record.Info.Upgrade4Points))
		totalPoints := fmt.Sprintf("%f", record.Info.TotalPoints)
		p1VoteScore := strconv.Itoa(int(record.Info.Proposal1VoteScore))
		p2VoteScore := strconv.Itoa(int(record.Info.Proposal2VoteScore))
		p3VoteScore := strconv.Itoa(int(record.Info.Proposal1VoteScore))
		p4VoteScore := strconv.Itoa(int(record.Info.Proposal2VoteScore))
		genPoints := strconv.Itoa(int(record.Info.GenesisPoints))
		addrObj := []string{address, record.Info.Moniker, uptimeCount, uptimePoints, up1Points,
			up2Points, up3Points, up4Points, p1VoteScore, p2VoteScore, p3VoteScore,
			p4VoteScore, genPoints, totalPoints}
		err := writer.Write(addrObj)

		if err != nil {
			log.Fatal("Cannot write to file", err)
		}
	}
}
