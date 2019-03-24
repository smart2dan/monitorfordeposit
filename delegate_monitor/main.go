package main

import (
	//"errors"
	"bufio"
	"delegate_monitor/db"
	"fmt"
	logger "github.com/alecthomas/log4go"
	"io"
	"os"
	"reflect"
	"runtime"
	"time"
)

const (
	defaultConfig = "./conf/usdtsyncd.conf"
	logConfig     = "./conf/log.xml"
)

var (
	helpTxt string
	mgo     *db.DbConn
	arrAddr []string
)

func Init() (*Configure, error) {
	runtime.GOMAXPROCS(4)
	cfg, err := LoadConfigure(defaultConfig)
	if err != nil {
		fmt.Println("config error.")
		return nil, err
	}
	logger.LoadConfiguration(logConfig)
	logger.Info("omniwallet command running!")
	return cfg, nil
}
func Contain(obj interface{}, target interface{}) bool {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true
		}
	}

	return false
}

func SendtoDelegate(c *WalletClient, delegate []string) {
	node_address := []string{"1HqqYeyw6xrvYANd6E8WVmbGHgXzjfeEaa",
		"1L4gUQjww9Z5j36rM4cz7cxCK9y1XDLLEs",
		"16TRCf2qbm1KAJUkwNQXsoUruRun6DbB99",
		"1NG6xarDqVnTUt8bTUgeGnJTWTYm1rr37J",
		"1BeFPKWdSkV6eudb1MbAQpcJKjTgFiAgz5",
		"1PdBtbS7UEMxmsVvZKV77bCZumyrzFnH2L",
		"1QEqBs2jr8vmk5HMYBxkQLisAyEbW1WgKQ",
		"19PgA6EG9AHQrANtcddu9L5aRk8JzxhTYq",
		"1Bfg8UqZbqdfXw2Bprn6iEuzTBpTb7cPH9",
		"1PbeXQncwtW9jVYsZXjP2jtzhh4BM8o5Wc",
		"1HRNKL5GXH1FLKQ3HsVNY9Ts6DKvDLQ2rZ",
		"1JVDHv2JWiVbSubQcRVSzrimZhpFJDhHxA",
		"1Fs92qAr93VEtQjWGCXU5rBo4XovkBJZdg",
		"1NRMYfAafn7Lv3TJ6RExJezGnTFBws9Qyd",
		"12EwcUTRbnauNbaernM2P1eS9HPFMT3VQp",
		"1J6bAS1zvPqb6SF1StfkF3BMGTo2TtWjNe",
		"1BHJnBSJHhhTevtkRBYmZtdkHeNS58fYWb",
		"1JogzL25R7DXptXTjCYAg1A1Kmni22F9He"}
	bank_addr, _ := c.GetAddrInfo()
	bank_addr_index := 0
	for _, v := range node_address {
		bv, _ := c.getBalance(v)
		balance := float64(bv) / 100000000

		for balance < 5000 {
			fmt.Println("Send fund to ", v)
			for ; bank_addr_index < len(bank_addr); bank_addr_index++ {
				if bank_addr[bank_addr_index].balance > 0.101 {
					fmt.Println(bank_addr[bank_addr_index].addr)
					c.SendFrom(bank_addr[bank_addr_index].addr, v, bank_addr[bank_addr_index].balance-0.101)
					time.Sleep(6 * time.Second)
					c.VoteforDelegate(bank_addr[bank_addr_index].addr, delegate, false)
					time.Sleep(6 * time.Second)
				}
				bv, _ = c.getBalance(v)
				balance = float64(bv) / 100000000
				if balance >= 5000 {
					goto Loop
				}
			}

		}
	Loop:
		fmt.Println(v, " balance is: ", balance)
		fmt.Println("next address")
	}
}
func VoteFor(c *WalletClient, delegate []string) {
	bank_addr, _ := c.GetAddrInfo()
	for _, v := range bank_addr {
		old_vote, _ := c.GetVoted(v.addr)
		if len(old_vote) == 0 {
			c.VoteforDelegate(v.addr, delegate, true)
			time.Sleep(6 * time.Second)
		} else if len(old_vote) == 51 {
			fmt.Println(v.addr, " already vote fund delegate!")
		} else {
			c.VoteforDelegate(v.addr, old_vote, false)
			time.Sleep(6 * time.Second)
			c.VoteforDelegate(v.addr, delegate, true)
			time.Sleep(6 * time.Second)
		}
	}
}
func ListVoteInfo(c *WalletClient) {
	f1, err := os.OpenFile("pool_vote_addr_1", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("os OpenFile error: ", err)
		return
	}
	defer f1.Close()
	f2, err := os.OpenFile("fund_vote_addr_1", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("os OpenFile error: ", err)
		return
	}
	defer f2.Close()
	bank_addr, _ := c.GetAddrInfo()
	for _, v := range bank_addr {
		old_vote, _ := c.GetVoted(v.addr)
		//fmt.Println(v.addr, "vote for", old_vote)
		if len(old_vote) < 49 {
			f1.WriteString(v.addr)
			f1.WriteString("\n")
		} else {
			f2.WriteString(v.addr)
			f2.WriteString("\n")
		}
	}

}
func ScanInfo(c *WalletClient) {

	bank_addr, _ := c.GetAddrInfo()
	total := float64(0.0)
	for _, v := range bank_addr {
		fmt.Println(v.addr+"  ", v.balance)
		total = total + v.balance
	}
	fmt.Println("total balance ", total)

}
func transfer_to_pool(c *WalletClient) {

	fi, err := os.Open("fund-userd.address")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer fi.Close()
	fo, err := os.Open("fund_vote_addr_1")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer fo.Close()

	br := bufio.NewReader(fi)
	br1 := bufio.NewReader(fo)
	for {
		addr, _, flag := br.ReadLine()
		if flag == io.EOF {
			break
		}
		addr_in, _, flag := br1.ReadLine()
		if flag == io.EOF {
			break
		}
		v, _ := c.getBalance(string(addr_in))
		amount := float64(v) / 100000000
		if amount > 0.02 {
			c.SendFrom(string(addr_in), string(addr), amount-0.02)
		} else {
			fmt.Println(string(addr_in))
			fmt.Println("not enough")

		}

		time.Sleep(1 * time.Second)
	}

}

func transfer_to_pool_1(c *WalletClient) {

	fi, err := os.Open("pool-used.address")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer fi.Close()
	fo, err := os.Open("pool_vote_addr")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer fo.Close()

	br := bufio.NewReader(fi)
	br1 := bufio.NewReader(fo)
	for {
		_, _, flag := br.ReadLine()
		if flag == io.EOF {
			break
		}
		addr_in, _, flag := br1.ReadLine()
		if flag == io.EOF {
			break
		}
		v, _ := c.getBalance(string(addr_in))
		amount := float64(v) / 100000000
		if amount > 0.02 {
			//c.SendFrom(string(addr_in), string(addr), amount-0.02)
			fmt.Println(string(addr_in))
		}
		// } else {
		// 	//fmt.Println(string(addr_in))
		// 	//fmt.Println("not enough")

		// }

		//time.Sleep(1 * time.Second)
	}

}
func laundering(c *WalletClient) {

}
func main() {

	// delegate_benpool := []string{"happy", "Cardner", "xterm", "drugstar", "FiggyLBTC", "BTCWolf", "LBTCNODE",
	// 	"LBTCGeek", "SuZhouLBTCer", "ChosenOne", "dhjdnu", "j", "chainAsia", "roelan", "BPC", "Shazam", "huahua",
	// 	"Ron", "Jacquelyn", "AnCao", "liuhua"}

	// delegate_fund := []string{"BTCC2", "BTCC1", "zthuning", "tothesky", "OUGCPbQ7", "yWuNeRdA", "Ydq2VTds", "EJJvtTZ7",
	// 	"p4j4864d", "quarkdai", "Zeyu", "tokenminer", "bixin", "victor", "bigStomachKing", "cai", "Dawei", "batbat",
	// 	"tokoyhot", "sprite", "chabei", "buymore", "ghostfaceuk", "drusilla", "seatrips", "trnpally", "ironman",
	// 	"allinlbtc", "richboss", "dreamer", "lovelbtc", "7fisher", "xiaomi", "bigminer", "catstar", "philhellmuth",
	// 	"purplegray", "24kman", "mababa", "desktop", "irine18", "teac", "batman", "Paladog", "omc", "HueyShin",
	// 	"smart2dan", "LBTCfans", "Lpineapple", "seven", "worldcoinguy"}

	cfg, err := Init()
	if err != nil {
		logger.Error("Init error: ", err)
		return
	}
	walletTask := CreateWalletTask(cfg)
	ScanInfo(walletTask)
	//ListVoteInfo(walletTask)
	//from := "1PuFpiUdAHeTTWtriaVxyRTJsAjzBQN8ej"
	//VoteFor(walletTask, delegate_benpool)
	//walletTask.getBalance(from)
	//SendtoDelegate(walletTask, delegate_fund)
	//mgo = db.CreateMongoClient(cfg.Db.Url, cfg.Db.DbName)
	//var people_addr []string
	/*
		fmt.Println(len(delegate_fund))
		walletTask := CreateWalletTask(cfg)
		all, err := walletTask.ListDelegates()
		hainan_addr, _ := walletTask.GetAddrInfo()
		fmt.Println(len(hainan_addr))
		count := 0
		for _, v := range hainan_addr {
			if count < 500 {
				if Contain(all[v.addr], delegate_fund) == false {
					fmt.Println(v.addr, all[v.addr])
					old_vote, _ := walletTask.GetVoted(v.addr)
					if len(old_vote) > 0 {
						walletTask.VoteforDelegate(v.addr, old_vote, false)
						time.Sleep(10 * time.Second)
					}
					walletTask.VoteforDelegate(v.addr, delegate_fund, true)
					time.Sleep(10 * time.Second)
				}
			}
			count = count + 1
		}
	*/

	// }
	// for _,v ：=range hainan_addr{
	// fmt.Println()
	// }
	//name, _ := walletTask.GetVoted("1Lx3sinZ7cw6tWRFe1MYoacBovYNGG6y1X")
	//walletTask.VoteforDelegate("1P8drBAwXoYPkjZSPhLaCDg9DaeQB5CoPy", name)
	//fmt.Println(string(addr))
	//walletTask.Conver()
	//mgo.CloseSession()
}
