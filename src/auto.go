package main

import (
	"net/http"
	"net/url"
	"fmt"
	"os"
	"strings"
	"strconv"
	"time"
	"math/rand"
	
	// xpath / html
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
	
	
	
)


func GetStateId(state string) int {

	switch state {
	case "Illinois":
		return 1;
	case "Michigan":
		return 2;
	case "California":
		return 3;
	case "New York":
		return 4;
	case "Nevada":
		return 8;
	case "New Jersey":
		return 9;
	default:
		fmt.Printf("Error state %s is not recognised!\n");
		os.Exit(1);
		return 0xff;	
	}
}

// pulls the timer from the response of a post request 
// to the auto page
func PullAutoRespTime(doc *html.Node) int {
	// pull auto manually for the next time to commit
	// <span class="countdown" data-length="90" data-start-time="1552331738" style="font-size:16px;">00:01:30</span><br><br>
		
	xpath := htmlquery.FindOne(doc, "//span[@class='countdown'][@data-length][@data-start-time][@style]");
		
	if(xpath == nil) {
		fmt.Println("Failed to pull auto resp time\n");
		os.Exit(1);
	}
		
	timer_str := htmlquery.SelectAttr(xpath, "data-length");
		
	sec, err := strconv.Atoi(timer_str);
		
	if(err != nil) {
		fmt.Println("Failed to convert to dec for auto resp!\n");
		os.Exit(1);
	}
	
	return sec;
}


func GetAutoChanceInt(str string) int {
	i := strings.Index(str, "%");
	
	if(i < 0) {
		fmt.Println("Unable to pull auto chance");
		os.Exit(1);
	}
	
	str = str[0:i];
	
	
	i, err := strconv.Atoi(str);
	
	if(err != nil) {
		fmt.Println("unable to convert auto chance string to int!");
		os.Exit(1);
	}
	
	return i;
}

// find max % if some are equal put in an array 
// rand out an answer
func GetMaxAutoSteal(doc *html.Node) int {

	arr := []int{0,0,0,0};

	
	// pull every chance into an array
	for i := 1; i <= 4; i++ {
		// table is offset by one from the id
		xpath_str := fmt.Sprintf("/html/body/table/tbody/tr[3]/td[2]/table/tbody/tr/td/div/form[1]/table/tbody/tr[%d]/td[2]",i+1);
		xpath := htmlquery.FindOne(doc,xpath_str);
		
		if(xpath == nil) {
			fmt.Printf("unable to pull auto timer for id: %d\n",i);
			os.Exit(1);
		}
		
		arr[i-1] = GetAutoChanceInt(htmlquery.InnerText(xpath));
	}
	
	// find the max value
	max := 0;
	max_idx := 0;
	
	for i := 0; i < len(arr); i++ {
		x := arr[i];
		if(x >= max) {
			max_idx = i+1;
			max = x;
		}
	}
	
	return max_idx;
}

// should repull timer from post repsonse
// but we will pull from the bar for now
func DoAuto(client *http.Client,config *Config) bool {

	// not read or not enabled
	if(!config.auto_enable || config.timers[TIMER_AUT] > 0) {
		return false;
	}

	
	
	config.no_actions++;

	// we need to pull the i value along with
	resp := SendGetReq(client,"https://www.bootleggers.us/autoburglary.php");
	doc, err := htmlquery.Parse(strings.NewReader(string(resp)));
	if(err != nil) {
		fmt.Println("Failed to parse the html for xpath");
		os.Exit(1);
	}
	
	// test for the capacha
	resp = TestCapacha(client,config,doc,resp,"https://bootleggers.us/autoburglary.php");
	
	// if in jail sleep it off then repull the page
	if(CheckJail(doc,config)) {
		resp = SendGetReq(client,"https://www.bootleggers.us/autoburglary.php");
		doc, err = htmlquery.Parse(strings.NewReader(string(resp)));
		if(err != nil) {
			fmt.Println("Failed to parse the html for xpath");
			os.Exit(1);
		}	
	}
	
	// Pull out our i value 
	i_val := GetIValue(doc);
	
	fmt.Printf("auto i value: %s\n",i_val);
	
	// <input type=radio name=select_crime_2944479 value="2"> Steal from a private parking lot</label>
	//                          ^ get this similar to the crime one
	xpath := htmlquery.FindOne(doc, "//input[@type='radio'][@name][@value]");
	sel_crime := htmlquery.SelectAttr(xpath,"name");
	fmt.Printf("%s\n",sel_crime);
		
	/* now we need to construct a request like so 
		select_crime_6696446	2 // number after is for private parking lot etc 
		steal_crew	# // if not from a crew
		i	a66964ae46fd44f57194080b11152476 // hidden i 
		shootCarThieves	1
	*/
	
	 // sleep for steal ( learn to audit code ;) )
	bot_sleep(4,4,config);
	
	
	
	
	
	// need to make it pick the auto value based on which one has the highest chance 
	// if they are equal pick at random
	// TODO
	
	value := GetMaxAutoSteal(doc);
	
	fmt.Printf("Max auto steal %d\n",value);
	
	
	// now send the request after this we need to check if the car was stolen
	// and then ship it to a random state that isnt ours
	resp = SendPostReq(client,"https://www.bootleggers.us/autoburglary.php",url.Values{sel_crime: {strconv.Itoa(value)},"steal_crew": {"#"},"i": {i_val},"shootCarThieves": {"1"}});
	doc, err = htmlquery.Parse(strings.NewReader(string(resp)));
	if(err != nil) {
		fmt.Println("Failed to parse the html for xpath");
		os.Exit(1);
	}
		
	// if we got caught no cars to ship so just return	
	if(strings.Contains(string(resp),"You are now in jail!")) { // in jail (can cause negative timers but with 60 sec > crime timer seems reasonable)
		
		sec := PullAutoRespTime(doc);
		
		config.timers[TIMER_AUT] = sec;	
	
	
		// pull jail page and update timers
		// subtract timers with the amount of time spent in jail
		resp := SendGetReq(client,"https://www.bootleggers.us/jail.php");
		doc, err = htmlquery.Parse(strings.NewReader(string(resp)));
		if(err != nil) {
			fmt.Println("Failed to parse the html for xpath");
			os.Exit(1);
		}	
		
		// update rac and crimes from bar
		config.timers[TIMER_CRI] = PullTimer(doc,"timer-cri");
		config.timers[TIMER_RAC] = PullTimer(doc,"timer-rac");
		config.timers[TIMER_TRAVEL] = PullTimer(doc,"timer-tra");
		
		// we know we are in jail so just sleep
		jail_sleep := PullTimer(doc,"timer-jai");
		
		// log jail
		dt := time.Now();
		str := fmt.Sprintf("[%s] in jail for %d",dt.String(),jail_sleep);
		log(str);
		
		
		fmt.Printf("In jail sleep for %d\n",jail_sleep);
		bot_sleep(jail_sleep,10,config);		
		return true;
	}
	
	
	// now look if we have stolen a car (both states the same)
	// then pull the plate and try and ship it
	
	


	// pull both states 
	// origin
	xpath = htmlquery.FindOne(doc,"/html/body/table/tbody/tr[3]/td[2]/table/tbody/tr/td/div/form[2]/table/tbody/tr[4]/td[5]");
	

	// need to check nil first to see if it could not find a xpath at all (i.e there are no cars)
	if(xpath == nil) {
	
		// update the timers
		doc, err = htmlquery.Parse(strings.NewReader(string(resp)));
		if(err != nil) {
			fmt.Println("Failed to parse the html for xpath");
			os.Exit(1);
		}
		
		// update rac and crimes from bar
		config.timers[TIMER_CRI] = PullTimer(doc,"timer-cri");
		config.timers[TIMER_RAC] = PullTimer(doc,"timer-rac");
		config.timers[TIMER_TRAVEL] = PullTimer(doc,"timer-tra");
		
		sec := PullAutoRespTime(doc);
		
		config.timers[TIMER_AUT] = sec;
		return true;
	}
	
	state_origin := htmlquery.InnerText(xpath);
	
 
	// current
	xpath = htmlquery.FindOne(doc,"/html/body/table/tbody/tr[3]/td[2]/table/tbody/tr/td/div/form[2]/table/tbody/tr[4]/td[6]");
	current_state := htmlquery.InnerText(xpath);
		

		
	// nothing to do as it does not need shipping (if its in a different state or in transit)
	if(current_state != state_origin || len(current_state) >= 10 && current_state[0:10] == "In transit") { // without a >= 10 check it could crash accessing the slice
		// update the timers ( can probably factor this away as its repeated twice)
		doc, err = htmlquery.Parse(strings.NewReader(string(resp)));
		if(err != nil) {
			fmt.Println("Failed to parse the html for xpath");
			os.Exit(1);
		}
		
		// update rac and crimes from bar
		config.timers[TIMER_CRI] = PullTimer(doc,"timer-cri");
		config.timers[TIMER_RAC] = PullTimer(doc,"timer-rac");
		config.timers[TIMER_TRAVEL] = PullTimer(doc,"timer-tra");
		
		sec := PullAutoRespTime(doc);

		config.timers[TIMER_AUT] = sec;
		return true;
	}
		
	// pull the plate
	//form[2]/table/tr[4]/td[2] // this could be better :P
	xpath = htmlquery.FindOne(doc,"/html/body/table/tbody/tr[3]/td[2]/table/tbody/tr/td/div/form[2]/table/tbody/tr[4]/td[2]");
	//fmt.Println(htmlquery.InnerText(xpath));

	plate := htmlquery.InnerText(xpath);
	fmt.Printf("plate: %s\n",plate);
	
	// finally ship it 
	
	// pick a random state to ship to
	// we have all ready pulled the origin so we can get the id of current state
	
	// const arrays aernt a thing in go 
	// and cause bsf didn't reindex the states when removing some 
	// there are gaps.... so use a switch statement
	
	id := GetStateId(current_state);
	

	
	
	// build a 6 array of ids
	// delete the element with our id of current sate
	// select from it randomly 0-5
	// to pull our shipping id
	a := [6]int{1,2,3,4,8,9};
	
	
	var id_str string;
	
	// if ship state is rand or current state is ship state
	// decide randomly
	if(config.ship_state == STATE_RAND || id == config.ship_state) {
		for i := 0; i < 6; i++ {
			if(a[i] == id) { // "delete" the element by swapping it to last
				temp := a[i];
				a[i] = a[5];
				a[5] = temp;
			}
		}
		
		// rand a id and convert to string
		// last id (a[5]) will contain our current state
		// and wont be included in the rand range
		id_str = strconv.Itoa(a[rand.Intn(5)]);
	} else { // ship to the selected state
		id_str = strconv.Itoa(config.ship_state);
	}
	
	fmt.Printf("Shipping to %s!\n",id_str);
	
	bot_sleep(5,7,config); // sleep for shipping
	
	// finally send off the ship request
	resp = SendPostReq(client,"https://www.bootleggers.us/autoburglary.php",url.Values{"shipPlate": {plate},"shipcar": {"Ship!"},"shipState": {id_str}});
	


	doc, err = htmlquery.Parse(strings.NewReader(string(resp)));
	if(err != nil) {
		fmt.Println("Failed to pull updated time value(auto)");
	}
	
	// have made an extra req can just update all the timers neatly
	
	UpdateTimers(doc,config);
	return true;
}