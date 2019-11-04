package main

import (
	"net/http"
	"net/url"
	"fmt"
	"os"
	"strings"
	"time"
	//"math/rand"
	
	// xpath / html
	"github.com/antchfx/htmlquery"
	
)


// all our do functions need jail checks	


// commit a crime may be better to return a new htmlnode
// than the timer as it will give a more updated version of 
// other timers
func DoCrime(client *http.Client,config *Config) bool {

	// not ready yet
	if(config.timers[TIMER_CRI] > 0) {
		return false;
	}


	// inc the number of actions done
	// we will do an extended sleep when we exceed
	// a certain value
	config.no_actions++;

	

	// pull the crime page
	resp := SendGetReq(client,"https://bootleggers.us/crimes.php");
	doc, err := htmlquery.Parse(strings.NewReader(string(resp)));
	if(err != nil) {
		fmt.Println("Failed to parse the html for xpath");
		os.Exit(1);
	}		

	bot_sleep(4,4,config); // sleep for commiting it
	
	
	// check for a capacha & jail before doing the crime
	resp = TestCapacha(client,config,doc,resp,"https://bootleggers.us/crimes.php");
	
	// if we are in jail we need to repull the crime page
	if(CheckJail(doc,config)) {
		resp = SendGetReq(client,"https://bootleggers.us/crimes.php");

		doc, err = htmlquery.Parse(strings.NewReader(string(resp)));
		if(err != nil) {
			fmt.Println("Failed to parse the html for xpath");
			os.Exit(1);
		}		
	}
	
	// pull the i value for the crime
	i := GetIValue(doc);
	
	fmt.Printf("crime i: %s\n",i);
	// pull business id and the number
	//<input type="radio" name="business_6291183" value="6">
	xpath := htmlquery.FindOne(doc, "//input[@type='radio'][@name][@value]");
	//fmt.Println(xpath);
	business := htmlquery.SelectAttr(xpath,"name")
	value := htmlquery.SelectAttr(xpath,"value")
	
	if(i == "") {
		fmt.Println("I value for crimes empty!?");
		os.Exit(1);
	} else if(value == "" || business == "") {
		fmt.Printf("%s:%s value or business is empty!?",value,business);
		os.Exit(1);
	}
	
	fmt.Printf("Crime done %s:%s!\n",business,value);
	
	// No use for response atm as it does not pull the updated timers bar
	// pull it when we want to print success or failure
	resp = SendPostReq(client,"https://www.bootleggers.us/crimes.php", url.Values{
		business: {value},
		"i": {i},
	});

	doc, err = htmlquery.Parse(strings.NewReader(string(resp)));
	if(err != nil) {
		fmt.Println("Failed to parse the html for xpath");
		os.Exit(1);
	}
 	
	
	// check the crime didnt fail if it did just pull the jail page and do a full timer update
	if(strings.Contains(resp,"You failed the crime and were arrested by police")) {
		// pull jail page and update timers
		// subtract timers with the amount of time spent in jail
		resp := SendGetReq(client,"https://www.bootleggers.us/jail.php");
		doc, err = htmlquery.Parse(strings.NewReader(string(resp)));
		if(err != nil) {
			fmt.Println("Failed to parse the html for xpath");
			os.Exit(1);
		}	
		UpdateTimers(doc,config);
		
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
	

	
	// Pull rackets and auto straight from the page
	config.timers[TIMER_AUT] = PullTimer(doc,"timer-aut");
	config.timers[TIMER_RAC] = PullTimer(doc,"timer-rac");
	config.timers[TIMER_TRAVEL] = PullTimer(doc,"timer-tra");
	
	// crime will have to be pulled specially from the resp
	xpath = htmlquery.FindOne(doc,"//span[@class='countdown-timeleft'][@style]");
	
	if xpath == nil {
		fmt.Println("Failed to pull updated crime timer from post req");
		os.Exit(1);
	}
	
	// get our str
	timer_str := htmlquery.InnerText(xpath);

	// pase the string for the int value in seconds
	config.timers[TIMER_CRI] = ParseRespTimer(timer_str);
	return true;
}