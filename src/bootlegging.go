package main

import (
	"net/http"
	"net/url"
	"fmt"
	"os"
	"strings"
	"regexp"
	"strconv"
	//"math/rand"
	
	// xpath / html
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
	
	
	
)


var bootleg_prices = [6][9] int {
	{40,98,200,303,416,614,960,1213,1776},
	{53,78,104,263,285,606,860,990,2121},
	{73,96,238,193,364,588,714,979,1746},
	{30,69,156,213,446,407,825,1045,1881},
	{40,100,167,294,402,549,780,1103,2200},
	{34,74,216,216,418,657,931,966,1665},
};

var bootleg_state = [6] string {"Michigan","Illinois","California","New York","Nevada","New Jersey"}; 	

type BootlegBuy struct {
    booze_id int;
    state_id int;
}




func GetBootleggingCapacity(resp string) int {
	x := strings.Index(resp,"Your carry capacity:");
	
	if(x == -1) {
		fmt.Println("Cannot pull capacity off bootlegging page!");
		os.Exit(1);
	}
	
	
	str := resp[x:x+32];
	
	
	
	
	// now we can just scan for an int and get our capacity
	re := regexp.MustCompile("[0-9]+");
	
	capacity_str := re.FindString(str);

	if(capacity_str == "") {
		fmt.Println("Unable to regex bootlegging capacity");
		os.Exit(1);
	}

	
	capacity, err := strconv.Atoi(capacity_str);
	
	if(err != nil) {
		fmt.Println("Could not convert bootlegging capacity to int");
		os.Exit(1);
	}
	return capacity;
}

// convets strings like "$12,355" to an int
func CashToInt(cash_str string) int {
	cash_str = cash_str[1:];
	cash_str = strings.Replace(cash_str,",","",-1); // remove commas
	
	
	cash, err := strconv.Atoi(cash_str);
	
	if(err != nil) {
		fmt.Printf("Could not convert cash to int(%s)\n",cash_str);
		os.Exit(1);
	}	
	
	return cash;
}

func TravelTo(client *http.Client,config *Config,state_id int) {


	// not ready
	if(config.timers[TIMER_TRAVEL] > 0) {
		fmt.Println("Not ready to travel!");
		return;
	}

	// pull the crime page
	resp := SendGetReq(client,"https://bootleggers.us/trainstation.php");
	doc, err := htmlquery.Parse(strings.NewReader(string(resp)));
	if(err != nil) {
		fmt.Println("Failed to parse the html for xpath");
		os.Exit(1);
	}		

	fmt.Println("Sleeping for travel time!");
	
	// sleep for picking the travel time
	bot_sleep(4,4,config);
	
	
	// check for a capacha before doing trying to travel :P
	resp = TestCapacha(client,config,doc,resp,"https://bootleggers.us/trainstation.php");
	
	// if in jail sleep it off then repull the page
	if(CheckJail(doc,config)) {
		resp = SendGetReq(client,"https://www.bootleggers.us/trainstation.php");
		doc, err = htmlquery.Parse(strings.NewReader(string(resp)));
		if(err != nil) {
			fmt.Println("Failed to parse the html for xpath");
			os.Exit(1);
		}	
	}
	
	fmt.Printf("Travelling to %d!\n",state_id);
	
	// send the travel req 
	resp = SendPostReq(client,"https://www.bootleggers.us/trainstation.php",url.Values{
		"travelto": {strconv.Itoa(state_id)},
	});
	
	
	// repull timers 
	doc, err = htmlquery.Parse(strings.NewReader(string(resp)));
	if(err != nil) {
		fmt.Println("Failed to pull updated time value(to_travel)");
	}
	
	UpdateTimers(doc,config);	
	// the timer will be too long to worry about so we will get it on the next 
	// timer repull as it wont update correctly after the first travel...
	config.timers[TIMER_TRAVEL] = 2147483647; // mark as not able to do twice
}

// get ammount of bought crates for each booze
func GetCrateAmmount(doc *html.Node) [9] int {
	var crate_ammounts [9] int;
	
	for i := 0; i < 9; i++ {
		crate_xpath := fmt.Sprintf("/html/body/table/tbody/tr[3]/td[2]/table/tbody/tr/td/div/table/tbody/tr[7]/td[%d]/span",i+1);
		
		xpath := htmlquery.FindOne(doc,crate_xpath);
		
		if(xpath == nil) {
			fmt.Println("Unable to pull crate ammount!");
			os.Exit(1);
		}
		
		crate_str := htmlquery.InnerText(xpath);
		
		if(crate_str == "") {
			fmt.Println("crate string is empty!");
			os.Exit(1);
		}
		
		cn, err := strconv.Atoi(crate_str);
		
		if(err != nil) {
			fmt.Println("Error converting crate str to int!");
			os.Exit(1);
		}
		crate_ammounts[i] = cn;
	}
	return crate_ammounts;
}


func FindBestBooze(state_idx int) BootlegBuy {
	// next we need to find the max profit booze in the current state 
	// and then calc how much of it we can buy and then put the request together
	
	max_dif := 0;
	var bootleg_buy BootlegBuy;
	
	// basically we will just search the table and find the max difference
	for i := 0; i < 9; i++ {
		buy := bootleg_prices[state_idx][i]; // buying from src
		var max int = 0;	
		for j := 0; j < 6; j++ { // go through every state and find the max for this booze
		
			// cant ship to same state we are in
			if(j == state_idx) {
				continue;
			}

			// max price
			if(bootleg_prices[j][i] > max) {
				max = bootleg_prices[j][i];
				
				// only care about the dest idx
				// as the booze id for selling and buying is the same 
				// and the current state is implied we only need to know the dest 
				// to travel there
				if((max - buy) > max_dif) {
					max_dif = max - buy;
					bootleg_buy.booze_id = i; 
					bootleg_buy.state_id = j;
				}				

			}			
			
		}

	}
	
	
	fmt.Printf("Max dif: %d, (%d), (%s->%s) (%d->%d)\n",
	max_dif,bootleg_buy.booze_id,bootleg_state[state_idx],bootleg_state[bootleg_buy.state_id],
	bootleg_prices[state_idx][bootleg_buy.booze_id],
	bootleg_prices[bootleg_buy.state_id][bootleg_buy.booze_id]);
	
	return bootleg_buy;
	
}

// pull an idx into our lookup table for the state
func GetBootlegStateIdx(state string) int {
	

	state_idx := 0;
	
	// now we have the state we need to convert it to an id into our bootleg table
	for i := 0; i < len(bootleg_state); i++ {
		if(bootleg_state[i] == state) {
			state_idx = i;
			break;
		}
	}
	
	if(state_idx == -1) {
		fmt.Printf("Unable to find state %s\n",state);
		os.Exit(1);
	}
	
	
	return state_idx;
}


func GetBoozePrice(id int, doc *html.Node) int {
	// table id is one higher than our internal one
	price_xpath := fmt.Sprintf("/html/body/table/tbody/tr[3]/td[2]/table/tbody/tr/td/div/table/tbody/tr[4]/td[%d]/span",id+1);
	

	xpath := htmlquery.FindOne(doc,price_xpath);
	
	if(xpath == nil) {
		fmt.Println("Unable to pull bootlegging prices!");
		os.Exit(1);
	}
		
	
	price_str := htmlquery.InnerText(xpath);
	
	if(price_str == "") {
		fmt.Println("bootlegging price is an empty string!");
		os.Exit(1);
	}
	
	price := CashToInt(price_str);
	
	
	return price;
}

// gets current player cash (may only work on the bootlegging page...)
func GetPlayerCash(doc *html.Node) int {
	xpath := htmlquery.FindOne(doc,"/html/body/table/tbody/tr[3]/td[3]/table[2]/tbody/tr[2]/td/table/tbody/tr/td/a[1]/div/p");
	
	if(xpath == nil) {
		fmt.Println("Unable to execute xpath for pulling player cash!");
		os.Exit(1);
	}
	
	player_cash_str := htmlquery.InnerText(xpath);
	
	
	if(player_cash_str == "") {
		fmt.Println("player cash is an empty string!");
		os.Exit(1);
	}
	

	
	player_cash :=  CashToInt(player_cash_str);
	return player_cash;
}

func BuyCrates(client *http.Client,config *Config,doc *html.Node, resp string) {
	// first we need to pull our current state
	xpath := htmlquery.FindOne(doc,"/html/body/table/tbody/tr[3]/td[3]/table[1]/tbody/tr[2]/td/table/tbody/tr[2]/td/div");

	if(xpath == nil) {
		fmt.Println("Error could not pull state off bootlegging page!");
		os.Exit(1);
	}
		
	state := htmlquery.InnerText(xpath);
	state_idx := GetBootlegStateIdx(state);


	// find out what booze we need to buy and where
	bootleg_buy := FindBestBooze(state_idx);
	
	
	
	// finally buy the booze & travel

	
	// first pull our capacity 
	// then the actual price and determine how much we can buy up to that capcity
	capacity := GetBootleggingCapacity(resp);
	
	

	
	// cool now we have to pull the price of the booze
	price := GetBoozePrice(bootleg_buy.booze_id,doc);

	
	
	
	// now we have to iter over the our current booze and find if for some shoddy reason we have any on us
	// it will later be checked above so we may just factor this shit out anywho
	
	crate_ammounts := GetCrateAmmount(doc);
	
	fmt.Println("crates: ", crate_ammounts);
	
	total_crates := 0;
	
	for _, x := range crate_ammounts {
		total_crates += x;
	}
	
	// and finally pull our cash and divide it by the price 
	// buy capacity if over or buy max we can afford
	player_cash := GetPlayerCash(doc);
	
	
	// now calc how many we can buy
	purch_ammount := 0;
	
	if((player_cash / price) > capacity) {
		purch_ammount = capacity;
	} else {
		purch_ammount = player_cash / price;
	}
	
	// minus the ammount we are buying off how many we allready have
	purch_ammount -= total_crates;
	
	if(purch_ammount == 0) {
		fmt.Println("No capacity for crates!?");
		
		// cool so by here we are just gonna have to travel to another state and dump it
		// need normal state id for travel (so lookup in our dest state in string array 
		// and pass to our normal id function
		TravelTo(client,config,GetStateId(bootleg_state[bootleg_buy.state_id]));		
		return;
	}
	
	fmt.Printf("Buying %d crates (%d,%d)\n",purch_ammount,player_cash,price);
	
	// now construct the req
	// purch[1]=&purch[2]=&purch[3]=&purch[4]=&purch[5]=&purch[6]=&purch[7]=&purch[8]=&purch[9]=&sell[1]=1&sell[2]=&sell[3]=&sell[4]=&sell[5]=&sell[6]=&sell[7]=&sell[8]=&sell[9]=
	// our request looks like the above so we basically just iter and add in the deets at the array index we need (the booze id) and we can buy things
	
	// need to figure out another way to build this request...
	req_str := "?";
	for i := 0; i < 8; i++ {
		table_id := i + 1;
		
		if(table_id == bootleg_buy.booze_id + 1) {
			req_str += fmt.Sprintf("purch[%d]=%d&",table_id,purch_ammount);
		} else {
			req_str += fmt.Sprintf("purch[%d]=&",table_id);
		}
	}
	
	if(9 == bootleg_buy.booze_id + 1) {
		req_str += fmt.Sprintf("purch[9]=%d&",purch_ammount);
	} else {
		req_str += "&purch[9]=";
	}
	  
	//and set all the sells to zero
	// yes this is a long ass string
	req_str += "&sell[1]=&sell[2]=&sell[3]=&sell[4]=&sell[5]=&sell[6]=&sell[7]=&sell[8]=&sell[9]=";
	
	req, err := url.ParseQuery(req_str);
	
	if(err != nil) {
		fmt.Println("Error encoding bootlegging buy request!");
		os.Exit(1);
	}
	
	fmt.Println("Sleeping for bootleg_buy");
	
	// sleep to give an abitary "buy delay"
	bot_sleep(4,4,config);
	
	resp = SendPostReq(client,"https://www.bootleggers.us/bootleg.php",req);
	
	
	// we have done a req so we can update our timers
	doc, err = htmlquery.Parse(strings.NewReader(string(resp)));
	if(err != nil) {
		fmt.Println("Failed to pull updated time value(buy_booze)");
	}
	
	UpdateTimers(doc,config);	
	
	
	// need normal state id for travel (so lookup in our dest state in string array 
	// and pass to our normal id function
	TravelTo(client,config,GetStateId(bootleg_state[bootleg_buy.state_id]));
}

func SellCrates(client *http.Client,config *Config,doc *html.Node, resp string) (string, *html.Node) {
	// dump any booze we have on us before we rebuy the next lot
	// we will check we have booze to sell first before we attempt this
	
	// so first things first pull our current inv 
	sell_crates := GetCrateAmmount(doc);
	fmt.Println("sell_crates: ",sell_crates);
	
	// sum the array and if we actually need to sell things else just return
	total_crates := 0;
	
	for _, x := range sell_crates {
		total_crates += x;
	}
	
	if(total_crates == 0) {
		fmt.Println("No crates to sell!");
		return resp, doc;
	}

	fmt.Println("Sleeping for bootleg_sell");
	bot_sleep(4,4,config);
	
	// now we know we need to sell crates all we need to do is build a query
	// we will sell each type individually to deal with crates being in the same state
	for i := 0; i < 9; i++ {
	
		// now we will dump a buy request only for the current one (if its zero we will ingore and just continue)

		if(sell_crates[i] == 0) {
			continue;
		}
	
		req_str := "?";

		
		req_str += "purch[1]=&purch[2]=&purch[3]=&purch[4]=&purch[5]=&purch[6]=&purch[7]=&purch[8]=&purch[9]=&";

		// build the sell part
		for j := 0; j < 8; j++ {
			table_id := j + 1;
			
			if(i == j) { // only want to sell one at a time (i.e one at current i value)
				req_str += fmt.Sprintf("sell[%d]=%d&",table_id,sell_crates[j]);
			} else {
				req_str += fmt.Sprintf("sell[%d]=&",table_id);
			}
		}
		
		if(sell_crates[8] != 0) {
			req_str += fmt.Sprintf("sell[9]=%d&",sell_crates[8]);
		} else {
			req_str += "&sell[9]=";
		}		
		
		
		req, err := url.ParseQuery(req_str);
		
		if(err != nil) {
			fmt.Println("Error encoding bootlegging buy request!");
			os.Exit(1);
		}

		resp = SendPostReq(client,"https://www.bootleggers.us/bootleg.php",req);
		
		// we have done a req so we can update our timers
		doc, err = htmlquery.Parse(strings.NewReader(string(resp)));
		if(err != nil) {
			fmt.Println("Failed to pull updated time value(buy_booze)");
		}
		
		fmt.Printf("Sell [%d] (%d)\n",i,sell_crates[i]);
		
		
		
		//if we manage to get back an error selling it for some reason
		// (if we have bought the booze in this state for some daft reason)
		// we will warn the user of this
		// * You can only sell booze that was purchased in another state! (find the string in the resp to confirm this)
		if(strings.Contains(resp,"* You can only sell booze that was purchased in another state!")) {
			fmt.Println("Warning attempted to sell booze bought in this state will try and sell in another!");
		}
	}

	// we have done a req so we can update our timers
	doc, err := htmlquery.Parse(strings.NewReader(string(resp)));
	if(err != nil) {
		fmt.Println("Failed to pull updated time value(buy_booze)");
	}

	UpdateTimers(doc,config);
	
	return resp, doc;
}
// perform a bootleg
func DoBootleg(client *http.Client,config *Config) bool {


	

	if(!config.bootleg_enable) {
		return false;
	}

	// not ready
	if(config.timers[TIMER_TRAVEL] > 0) {
		return false;
	}

	
	
	fmt.Println("Doing bootleg!");
	config.no_actions++;
	
	resp := SendGetReq(client,"https://www.bootleggers.us/bootleg.php");
	doc, err := htmlquery.Parse(strings.NewReader(string(resp)));
	if(err != nil) {
		fmt.Println("Failed to parse the html for xpath");
		os.Exit(1);
	}

	// check capacha before doing anything else
	resp = TestCapacha(client,config,doc,resp,"https://bootleggers.us/trainstation.php");

	// if in jail sleep it off then repull the page
	if(CheckJail(doc,config)) {
		resp = SendGetReq(client,"https://www.bootleggers.us/bootleg.php");
		doc, err = htmlquery.Parse(strings.NewReader(string(resp)));
		if(err != nil) {
			fmt.Println("Failed to parse the html for xpath");
			os.Exit(1);
		}	
	}


	

	
	resp, doc  = SellCrates(client,config,doc,resp);
	
	// we have done a req so we can update our timers
	doc, err = htmlquery.Parse(strings.NewReader(string(resp)));
	if(err != nil) {
		fmt.Println("Failed to pull updated time value(buy_booze)");
	}
	
	
	// buy a new load
	BuyCrates(client,config,doc,resp);
	
	
	return true;
}