package reward

import (
	"github.com/ethereum/go-ethereum/ethdb/leveldb"
	"log"
	"math/big"
	"os/user"
)

var recodDB *leveldb.Database

func NewRecordDB() *leveldb.Database {
	if recodDB != nil {
		return recodDB
	}
	var err error
	u, err := user.Current()
	if nil != err {
		log.Fatal("leveldb new error", err)
		return nil
	}
	recodDB, err = leveldb.New(u.HomeDir+"/.nuc/mining-rewards", 32, 32, "mining")
	if err != nil {
		log.Fatal("leveldb new error", err)
		return nil
	}
	return recodDB
}

func GetRewardsByNumber(number *big.Int) []byte {
	db := NewRecordDB()
	result, err := db.Get(number.Bytes())
	if err != nil {
		return []byte{}
	}
	return result
}
