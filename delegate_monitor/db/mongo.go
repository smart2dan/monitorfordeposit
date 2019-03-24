package db

import (
	logger "github.com/alecthomas/log4go"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	tsTable    = "transaction"
	addrTable  = "coin_address"
	lastRecord = "lastrecord"

	lastSyncRecordId = "FFFFFFFFFFFFFFFFFFFFFEEE"
)

var (
	LastRecordId string = "FFFFFFFFFFFFFFFFFFFFFEEE"
)

type DbConn struct {
	session *mgo.Session
	url     string
	dbName  string
	//lastRecordId string
}

type Transaction struct {
	Id               bson.ObjectId `bson:"_id"`
	Amount           string        `bson:"amount"`
	Block            int           `bson:"block"`
	Blockhash        string        `bson:"blockhash"`
	Blocktime        int64         `bson:"blocktime"`
	Confirmations    int           `bson:"confirmations"`
	Divisible        bool          `bson:"divisible"`
	Fee              string        `bson:"fee"`
	Ismine           bool          `bson:"ismine"`
	Positioninblock  int           `bson:"positioninblock"`
	Propertyid       int           `bson:"propertyid"`
	Referenceaddress string        `bson:"referenceaddress"`
	Sendingaddress   string        `bson:"sendingaddress"`
	Txid             string        `bson:"txid"`
	Type             string        `bson:"type"`
	Typeint          int           `bson:"typeint"`
	Valid            bool          `bson:"valid"`
	Version          int           `bson:"version"`
}

type SyncLastTrans struct {
	Id             bson.ObjectId `bson:"_id"`
	SerId          string        `bson:"serid"`
	Sendingaddress string        `bson:"sendingaddress"`
	Txid           string        `bson:"txid"`
	Block          int           `bson:"block"`
}

type CoinAddress struct {
	Id      bson.ObjectId `bson:"_id"`
	Address string        `bson:"address"`
}

/**
 * 公共方法，获取session，如果存在则拷贝一份
 */
func getConn(url string) *mgo.Session {
	session, err := mgo.Dial(url)
	if err != nil {
		logger.Error("connect mongodb failed! err: ", err)
		return nil
	}
	//session.SetMode(mgo.Monotonic, true)
	return session
}

func CreateMongoClient(url, dbName string) *DbConn {
	conn := &DbConn{
		url:     url,
		dbName:  dbName,
		session: getConn(url),
	}
	return conn
}

func (s *DbConn) CloseSession() {
	if s.session != nil {
		s.session.Close()
	}
}

func (s *DbConn) GetSession() *mgo.Session {
	if s.session == nil {
		s.session = getConn(s.url)
		if s.session == nil {
			return nil
		}
	}
	//最大连接池默认为4096
	return s.session.Clone()
}

//公共方法，获取collection对象
func (s *DbConn) WitchCollection(collection string, sc func(*mgo.Collection) error) error {
	session := s.GetSession()
	if session == nil {
		return nil
	}
	defer session.Close()
	c := session.DB(s.dbName).C(collection)
	if c == nil {
		logger.Error("select collection failed!")
		return nil
	}
	return sc(c)
}

func (s *DbConn) GetDbCollection(collection string) *mgo.Collection {
	session := s.GetSession()
	if session == nil {
		return nil
	}
	defer session.Close()
	c := session.DB(s.dbName).C(collection)
	if c == nil {
		logger.Error("select collection failed!")
		return nil
	}

	return c
}

func (s *DbConn) FindTransaction(addr string) ([]Transaction, error) {
	var ts []Transaction
	query := func(c *mgo.Collection) error {
		//return c.Find(bson.M{"$or": []bson.M{bson.M{"sendingaddress": addr}, bson.M{"referenceaddress": addr}}}).All(&ts)
		return c.Find(&bson.M{"sendingaddress": addr}).All(&ts)
	}
	err := s.WitchCollection(tsTable, query)
	if err != nil {
		return ts, err
	}
	return ts, nil
}

func (s *DbConn) AddTransaction(ts *Transaction) (string, error) {
	ts.Id = bson.NewObjectId()
	//logger.Info("id: %s address: %s tx: %s amount: %s blockheight: %d fee: %s confir: %d txtime: %d",
	//	ts.Id, ts.Sendingaddress, ts.Txid, ts.Amount, ts.Block, ts.Fee, ts.Confirmations, ts.Blocktime)
	query := func(c *mgo.Collection) error {
		return c.Insert(ts)
	}
	err := s.WitchCollection(tsTable, query)
	if err != nil {
		return "", err
	}
	return ts.Id.Hex(), nil
}

func (s *DbConn) RemoveTransaction(blackHeight int) error {
	query := func(c *mgo.Collection) error {
		return c.Remove(bson.M{"block": blackHeight})
	}
	err := s.WitchCollection(tsTable, query)
	if err != nil {
		return err
	}
	return nil
}

//记录最后一次交易信息， 重启服务后继续同步
func (s *DbConn) AddSyncLastTransInfo(sync *SyncLastTrans) (string, error) {
	sync.Id = bson.NewObjectId()
	query := func(c *mgo.Collection) error {
		return c.Insert(sync)
	}
	err := s.WitchCollection(lastRecord, query)
	if err != nil {
		return "", err
	}
	return sync.Id.Hex(), nil
}

func (s *DbConn) FindSyncLastTransInfo(serid string) (SyncLastTrans, error) {
	var sync SyncLastTrans
	query := func(c *mgo.Collection) error {
		return c.Find(&bson.M{"serid": serid}).One(&sync)
	}
	err := s.WitchCollection(lastRecord, query)
	if err != nil {
		return sync, err
	}
	return sync, nil
}

func (s *DbConn) UpdateSyncLastTransInfo(sync *SyncLastTrans) error {
	query := func(c *mgo.Collection) error {
		return c.Update(bson.M{"serid": sync.SerId},
			bson.M{"$set": bson.M{"block": sync.Block,
				"sendingaddress": sync.Sendingaddress, "txid": sync.Txid}})
	}
	err := s.WitchCollection(lastRecord, query)
	if err != nil {
		return err
	}
	return nil
}

func (s *DbConn) FindAddr(addr string) (string, error) {
	var coinAddr CoinAddress
	query := func(c *mgo.Collection) error {
		return c.Find(&bson.M{"sendingaddress": addr}).One(&coinAddr)
	}
	err := s.WitchCollection(addrTable, query)
	if err != nil {
		return "", err
	}
	return coinAddr.Address, err
}

func (s *DbConn) AddCoinAddr(addr string) (string, error) {
	var ca CoinAddress
	ca.Id = bson.NewObjectId()
	ca.Address = addr
	query := func(c *mgo.Collection) error {
		return c.Insert(ca)
	}
	err := s.WitchCollection(addrTable, query)
	if err != nil {
		return "", err
	}
	return ca.Id.Hex(), err
}
