package main

import (
	"net/http"
	"net/url"
	"fmt"
	"strings"
	"strconv"
	//"math/rand"
	
	// xpath / html
	"github.com/antchfx/htmlquery"
	
)



// do rackets can pull timer for json if we really need to
// also need to add capacha det from the json

// need delays between sends 

func DoRackets(client *http.Client, config *Config) bool {
	
	// not enabled or not ready return immediately 
	if(!config.racket_enable || config.timers[TIMER_RAC] > 0) {
		return false;
	}

	config.no_actions++;
	
	fmt.Println("Committed racket!");

	//At this point we know something is ready on the rackets page 
	// So we have to figure out what needs starting and what needs collecting
	// first do a get req to the page
	resp := SendGetReq(client, "https://www.bootleggers.us/rackets.php");
	
	// get it ready to parse
	doc, err := htmlquery.Parse(strings.NewReader(string(resp)));
	if(err != nil) {
		fmt.Println("Failed to parse rackets page");
	}
	
	// check we aint got a capacha or jail
	resp = TestCapacha(client,config,doc,resp,"https://bootleggers.us/rackets.php");
	injail := CheckJail(doc,config);
	

	
	
	
	// if we are just in jail the page is screwed so we need to reparse it
	if(injail) {
		resp := SendGetReq(client, "https://www.bootleggers.us/rackets.php");
		
		// get it ready to parse
		doc, err = htmlquery.Parse(strings.NewReader(string(resp)));
		if(err != nil) {
			fmt.Println("Failed to parse rackets page");
		}	
	}
	
	// do before the main loop as it will just have json
	// see below for ideal solution
	UpdateTimers(doc,config);	
	
	
	// check each id if its ready to collect
	// collect and start at once 
	// if its ready to start just start it
	// else do nothing
	// need jail check + capacha check after each action in the loop
	//</li><li class="crime BL-bg-dark" data-id="3" data-status="idle"> (commit this)
	//<li class="crime BL-bg-dark" data-id="1" data-status="collectable"> (collect this)
	for id := 1; id <= 4; id++ {		
		// put together our query
		id_str := strconv.Itoa(id); // convert our id to string
		query := "//li[@class='crime BL-bg-dark'][@data-id='" + id_str + "']" + "[@data-status]"
			
		xpath := htmlquery.FindOne(doc, query);
			
		// pull the data status and decide what to do
		data_status := htmlquery.SelectAttr(xpath,"data-status")
		fmt.Printf("%d: %s\n",id,data_status);
				
		if(data_status == "collectable") { // collect and start
			bot_sleep(4,4,config);
			resp = SendPostReq(client, "https://www.bootleggers.us/ajax/rackets.php?action=collect",url.Values{"id": {id_str}});
			fmt.Printf("json: %s\n",resp);
			bot_sleep(2,2,config); // mouse is in the same place so this action will take a substantal ammount less
			resp = SendPostReq(client, "https://www.bootleggers.us/ajax/rackets.php?action=start",url.Values{"id": {id_str}});
			fmt.Printf("json: %s\n",resp);
		} else if(data_status == "idle") { // start the racket
			bot_sleep(4,4,config); // sleep for commiting it
			resp = SendPostReq(client, "https://www.bootleggers.us/ajax/rackets.php?action=start",url.Values{"id": {id_str}});
			fmt.Printf("json: %s\n",resp);
		}	
	}


	// need to parse timers out from json ideally on this or off the page directly as rackets 
	// are special and have multiple timers
	// but in practice rackets will have the longest timer after we are done
	config.timers[TIMER_RAC] = 2147483647;
	return true;
}