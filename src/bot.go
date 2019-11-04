// add checks for nil to every htmlquery.FindOne with a better error message
// however still exit as if things are not as we expect pages may have been altered 




// missing timers due to jail times 
// after a jail submit every timer must be repulled
// should use a global array and keep them up to date to avoid
// inaccuracy also need to pull from the page timers directly


// page we pull for the loops initial timers should not be pulled from crimes as we can just subtract the sleep timer
// off every timer to get the next set beyond pulling the initial ones







// ANTI DECTECTION

// randomly view forums 

// randomize order of doing actions via array of function ptrs - done 
// add breaks  - done
// increase the variance in delay i.e do a large one or dont do it at all - done to an extent
// miss certain shorter actions i.e crimes on purpose ever so often - todo

// occasionally tick the 2nd longest timer and do both in one go (wait for both to elapse and do both at same time) - todo

// log off for an hour or two occasionally (even if we have fekk all time) - todo when its fully automated
// add a logger to dump all our post and get reqs out to a file with a timestamp
// so we can improve the bot


// add a  logger to the bot 
// also add random logouts and various other actions - todo
// like checking hte grave page forums etc

// also need it sometimes checking actions far too early on purpose - (done but needs improvement)


package main

import (
	"bytes"
	"net/http"
	"net/url"
	"net/http/cookiejar"
	//"golang.org/x/net/proxy"
	//"encoding/json"
	"fmt"
	"os"
	"io/ioutil"
	"io"
	"strings"
	"strconv"
	"time"
	"bufio"
	"math/rand"
	// xpath / html
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
	"github.com/gocolly/twocaptcha"
	
	"github.com/chromedp/chromedp"
	"context"

)

var user_agent string; // contains the user agent

// dummy config file 
var config_dummy string = "use_session_id=true\nusername=changeme\npassword=changeme\nsessid=changeme\nuser_agent=changeme\ndo_rackets=changeme\ndo_auto=changme\nship_state=changeme\ndo_bootleg=changme\nuser_cookie_val=changeme\napi_key=NULL";

const (
	TIMER_AUT = 0;
	TIMER_CRI = 1;
	TIMER_RAC = 2;
	TIMER_TRAVEL = 3;
	STATE_RAND = 0x1337;
	TIMER_RAND = 4; // random action (not a real timer)
)

type Config struct {
	racket_enable bool
	auto_enable bool
	bootleg_enable bool
	timers[5]int
	no_actions int
	ship_state int
	aut_early_time int
	crime_early_time int
	rackets_early_time int
	api_key string
}





func UpdateTimers(doc *html.Node, config *Config) {
	config.timers[TIMER_CRI] = PullTimer(doc,"timer-cri");
	config.timers[TIMER_AUT] = PullTimer(doc,"timer-aut");
	config.timers[TIMER_RAC] = PullTimer(doc,"timer-rac");
	config.timers[TIMER_TRAVEL] = PullTimer(doc,"timer-tra");
}







// wrap building a header and receiving a response for post and get

func SendPostReq(client *http.Client, url string, parameters url.Values) string {
	req, err := http.NewRequest("POST",url,strings.NewReader(parameters.Encode()));
	req.PostForm = parameters;
	if err != nil {
		fmt.Printf("Failed to build post request url for %s",url);
		os.Exit(1);			
	}
	// mask the user agent and setup the correct content type
	req.Header.Add("User-Agent", user_agent);
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(parameters.Encode()))) 
	 
	
	dt := time.Now();
	str := fmt.Sprintf("[%s] POST %s?%s",dt.String(),url,strings.NewReader(parameters.Encode()));
	log(str); 
	 
	// do post req
	resp, err := client.Do(req)
	if (err != nil) {
		fmt.Printf("Failed to send post req for %s are you connected to the internet?\n",url);
		os.Exit(1);
	}
	
	// read out response
	body, err := ioutil.ReadAll(resp.Body)
	if(err != nil) {
		fmt.Printf("Failed to read post response for %s",url );
		os.Exit(1);	
	}
	resp.Body.Close()

	// sleep for "load times"
	time.Sleep(time.Duration(rand.Float64()) * time.Second);
	return string(body);	
}



func SendGetReq(client *http.Client, url string) string {
	req , err := http.NewRequest("GET",url,nil);
	if err != nil {
		fmt.Printf("Failed to get request url for %s",url);
		os.Exit(1);			
	}
	
	// mask the user agent and setup the correct content type
	req.Header.Add("User-Agent", user_agent);	
	
	
	// log our get request!
	dt := time.Now();
	
	str := fmt.Sprintf("[%s] GET %s",dt.String(),url);
	log(str);
	
	// send our get req
	resp, err := client.Do(req)
	if(err != nil) {
		fmt.Printf("Failed to send get req for %s are you connected to the Internet?\n",url);
		os.Exit(1);	
	}
	
	// read out the response
	body, err := ioutil.ReadAll(resp.Body)
	if(err != nil) {
		fmt.Printf("Failed to read get response for %s",url );
		os.Exit(1);	
	}
	resp.Body.Close()

	// sleep for "load times"
	time.Sleep(time.Duration(rand.Float64()) * time.Second);
	return string(body);		
}




// not needed i think as bsf doesent have any json forms?

func SendPostReqJson(client *http.Client, url string, parameters []byte) string {


	req, err := http.NewRequest("POST",url,bytes.NewBuffer(parameters));
	if err != nil {
		fmt.Printf("Failed to build post request url for %s",url);
		os.Exit(1);			
	}
	// mask the user agent and setup the correct content type
	req.Header.Add("User-Agent", user_agent);
	req.Header.Add("Content-Type", "application/json")
	 
	// do post req
	resp, err := client.Do(req)
	if (err != nil) {
		fmt.Printf("Failed to send post req for %s\n",url);
		os.Exit(1);
	}
	
	// read out response
	body, err := ioutil.ReadAll(resp.Body)
	if(err != nil) {
		fmt.Printf("Failed to read post response for %s",url );
		os.Exit(1);	
	}
	resp.Body.Close()

	return string(body);
}






// pull the "i" value on the page 

func GetIValue(doc *html.Node) string {
/*	// try and pull i value from crimes with xpath
	body := SendGetReq(client,"https://bootleggers.us/crimes.php");

	
	
	// parse out the i value
	doc, err := htmlquery.Parse(strings.NewReader(string(body)));
	if(err != nil) {
		fmt.Println("Failed to pull i value");
	}
*/

	// ^ old code we pass in the most recently fetched page 
	// so we aint constantly refreshing crimes
	
	
	//<input type="hidden" name="i" value="62c9fc11f83d23721417f4b087c7aaf8">
	xpath := htmlquery.FindOne(doc, "//input[@type='hidden'][@name='i']");
	

	i := htmlquery.SelectAttr(xpath,"value");
	
	
	return i;
}


func VisitPageAndDie(client *http.Client) {
	 forum_link := "https://www.bootleggers.us/forum_new/index.php?flag="

	// 1 - 3 are game, offtopic, and classifields forum pages
	rand_num := rand.Intn(3) + 1; // random number between 3 and one
		
	// cat onto end of link
	forum_link += strconv.Itoa(rand_num);
		
	
	// visit a random forum to make the capacha look less odd (improve this to do something a little more advanced)
	SendGetReq(client,forum_link);
	
	fmt.Printf("Capatcha detected logging off!\n");
	os.Exit(1);
}


// test for a capacha just logout for now
func TestCapacha(client *http.Client,config *Config,doc *html.Node,resp string, link string) string{

	

	// if this is present then capacha is up 
	xpath := htmlquery.FindOne(doc, "//input[@type='submit'][@value='Continue playing!']");
	capacha := htmlquery.SelectAttr(xpath,"value")
		
	if(capacha == "Continue playing!")  {// detected capacha terminate bot (need to add solver)
	
		log("Captcha hit");
		if(config.api_key == "NULL") {
			VisitPageAndDie(client);
			return "DIE";
		} else { // incredibly broken code (not ready yet lol)
			xpath := htmlquery.FindOne(doc,"//div[@class='g-recaptcha']");
			sitekey := htmlquery.SelectAttr(xpath,"data-sitekey");
			fmt.Printf("capacha sitekey %s\n",sitekey);
		
			fmt.Printf("Sending capacha solve req for %s\n",link);
			/* hand rolled code
			// now we have the sitekey
			// we need to construct a solve request
			resp := SendPostReq(client,"https://2captcha.com/in.php",url.Values{
				"key": {config.api_key},
				"method": {"userrecaptcha"},
				"googlekey": {sitekey},
				"pageurl": {link},
			});
			fmt.Println(resp);
			
			// get back a resp like this
			// OK|62364970887
			if(resp[0:2] != "OK") {
				fmt.Println("An error occured pulling the capacha!");
				VisitPageAndDie(client);
			}
			
			fmt.Println("Sleeping for capacha response..");
			
			time.Sleep(time.Duration(21) *time.Second);
			
			req_url := fmt.Sprintf("https://2captcha.com/res.php?key=%s&action=get&id=%s",config.api_key,resp[3:]);
			
			fmt.Printf("Sending req: https://2captcha.com/res.php?key=REDACTED&action=get&id=%s\n",resp[3:]);
			
			for {
				resp := SendGetReq(client,req_url);
				if(resp == "CAPCHA_NOT_READY") {
					fmt.Println("capacha not ready!");
					time.Sleep(time.Duration(8) *time.Second);
				} else if(resp[0:5] == "ERROR") {
					fmt.Printf("Error solving capacha %s\n",resp);
					VisitPageAndDie(client);
				} else {
					fmt.Printf("Got key: %s\n",resp[3:]);
					break;
				}
			}
			
			
			
			
			
			// sleep before the send
			time.Sleep(time.Duration(1 + rand.Intn(2)) *time.Second);
			
			req_str := link + "?g-recaptcha-response=" + resp[3:];
			// finally send off the resulting key and update timers #
			*/
			
			c := twocaptcha.New(config.api_key);
			
			
			key, err := c.SolveRecaptchaV2(link,sitekey);
			
			if(err != nil) {
				fmt.Println("Error solving capacha!");
				os.Exit(1);
			}
			/*
			req_str := link + "?g-recaptcha-response=" + key;
			
			fmt.Printf("req: %s\n",req_str);
			
			req, err := url.ParseQuery(req_str);
			
			if(err != nil) {
				fmt.Println("Error encoding bootlegging buy request!");
				os.Exit(1);
			}			
			*/
			
			fmt.Println(key); 
			
			resp = SendPostReq(client,link,url.Values{"g-recaptcha-response": {key}}); 
			
			// update the doc so it can be reused as well as this return a string with the resp for good measure
			doc, err := htmlquery.Parse(strings.NewReader(string(resp)));
			if(err != nil) {
				fmt.Println("Failed to parse the html for xpath");
				os.Exit(1);
			}
			
			UpdateTimers(doc,config);
			
			// if this is present then capacha somehow still up 
			xpath = htmlquery.FindOne(doc, "//input[@type='submit'][@value='Continue playing!']");
			capacha = htmlquery.SelectAttr(xpath,"value")
			
			if(capacha == "Continue playing!") {
				fmt.Println("Capatcha still up after solve!?");
				log(resp);
				VisitPageAndDie(client);
			}
			
			return resp;
		}
		
	}
	return resp;
}


// parses a timer after we have committed a crime so 
// we can get the next time
func ParseRespTimer(timer_str string) int {
	// now we have a str like this 00:00:45
	// we have to convert it to seconds
	
	var seconds int = 0;
	
	tmp, err := strconv.Atoi(timer_str[0:2]); 
	
	if(err != nil) {
		fmt.Println("Failed to convert response timer(sec)");
		os.Exit(1);
	}
	
	seconds += (tmp * 60 * 60); // 60^2 seconds in an hour
	
	tmp, err = strconv.Atoi(timer_str[3:5]);
	if(err != nil) {
		fmt.Println("Failed to convert response timer(min)");
		os.Exit(1);
	}	
	seconds += (tmp  * 60); // 60 seconds in a min;
	
	tmp, err = strconv.Atoi(timer_str[6:8]);
	if(err != nil) {
		fmt.Println("Failed to convert response timer(hr)");
		os.Exit(1);
	}	
	seconds += tmp;
	
	return seconds
}










// returns a timer from the timers bar given its name on the page 
// e.g timer-cri

func PullTimer(doc *html.Node,timer_name string) int {
	
	// <span id="timer-aut" data-seconds="-1630524" style="color: #8EF393">Ready</span>
	timer_xpath := "//span[@id='" + timer_name + "']";
	
	// check the timer
	xpath := htmlquery.FindOne(doc, timer_xpath);
	timeleft_str := htmlquery.SelectAttr(xpath,"data-seconds")
		
	// get the time as int
	remaining, _ := strconv.Atoi(timeleft_str);
	
	
	return remaining;
}

// check jail and sleep it off
// make it check busts later
// return true if we were in it as we will have to repull
// the page (so our requests look legit)


// make this repull timer repeatedly 

func CheckJail(doc *html.Node,config *Config) bool {


	
	// check the jail timer
	xpath := htmlquery.FindOne(doc, "//span[@id='timer-jai']");
	timeleft_str := htmlquery.SelectAttr(xpath,"data-seconds")
		
	// get the time as int
	remaining, _ := strconv.Atoi(timeleft_str);

	// we are in jail so sleep it off
	if(remaining > 0) {
	
		// log jail
		dt := time.Now();
		str := fmt.Sprintf("[%s] in jail for %d",dt.String(),remaining);
		log(str);
		
		fmt.Printf("in jail sleeping for %d\n",remaining);
		bot_sleep(remaining,10,config);
		return true;
	}
	return false; // wernt in jail
}



func fileExists(filename string) bool {
    info, err := os.Stat(filename);
    if os.IsNotExist(err) {
        return false;
    }
    return !info.IsDir();
}

// wrapper to write a string to a file and make it
func WriteToFile(filename string, buf string) {
    fp, err := os.Create(filename);
    if err != nil {
		fmt.Printf("failed to create file: %v",err);
		os.Exit(1);
    }
    defer fp.Close();

    _, err = io.WriteString(fp, buf);
    if err != nil {
        fmt.Printf("failed to write string: %v",err);
		os.Exit(1);
    }
}


var rand_pages = [] string {"graveyard.php","usersonline.php","states.php","stats.php","crimes.php","autoburglary.php","rackets.php","forum_new/index.php","gold.php?page=6","mailbox.php"}; 	

// do a random action (to legitmize traffic)
// may need to make the timestep have a tad more variance
func DoRandom(client *http.Client,config *Config) {

	// not ready yet
	if(config.timers[TIMER_RAND] > 0) {
		return;
	}
	
	// pick a random page
	rand_page :=  "https://www.bootleggers.us/" + rand_pages[rand.Intn(len(rand_pages))];
	
	fmt.Printf("Randomly checking %s\n",rand_page);
	
	// pull the crime page (will be a random string)
	resp := SendGetReq(client,rand_page);
	doc, err := htmlquery.Parse(strings.NewReader(string(resp)));
	if(err != nil) {
		fmt.Println("Failed to parse the html for xpath");
		os.Exit(1);
	}		

	// sleep off "visting the page"
	bot_sleep(4,4,config);
	
	// check for a capacha just for saftey
	resp = TestCapacha(client,config,doc,resp,rand_page);
	

	// might as well update timers while here
	UpdateTimers(doc,config);
	
	// set a new random timer
	config.timers[TIMER_RAND] =  296 + rand.Intn(420);
	fmt.Printf("Next Random Action in %d\n",config.timers[TIMER_RAND]);
}


func GetMinTimer(config *Config) int {
	// assume cri is min
	min_timer := config.timers[TIMER_CRI];
		
	// enabled and min timer
	if(config.auto_enable && config.timers[TIMER_AUT] < min_timer) {
		min_timer = config.timers[TIMER_AUT];
	}
		
		// enabled and min timer
	if(config.racket_enable && config.timers[TIMER_RAC] < min_timer) {
		min_timer = config.timers[TIMER_RAC];
	}
		
	if(config.bootleg_enable && config.timers[TIMER_TRAVEL] < min_timer) {
		min_timer = config.timers[TIMER_TRAVEL];
	}
	return min_timer
}

// sleep for a time and subtract timers
func bot_sleep(min int,variance int, config *Config) {

	if variance == 0 { // variance of zero can panic the rand func
		panic("Zero variance for bot sleep!");
	}


	sleep_timer := float64(min) + rand_time(variance);

	if(sleep_timer < 2.0) { // need delays over 2 seconds otherwhise bot is operating too fast
		panic("Inhuman browsing rate");
	}
	
	time.Sleep(time.Duration(sleep_timer) * time.Second);
	
	// subtract the sleep time from all timers
	for i := 0; i< len(config.timers); i++ {
		config.timers[i] -= int(sleep_timer); // sleeping for this so for next iter we want that removed
	}
}


func rand_time(t int) float64 {
	x := float64(rand.Intn(t)) + rand.Float64();
	return x;
}

func main() {
	// Create a transport that uses Tor Browsers SocksPort
	
 
 
 
 
	
	fmt.Printf("Starting bot!...\n");
	
	// init our cookie jar + client
	jar, _ := cookiejar.New(nil);
	
	client := &http.Client{
		Jar: jar,
		//Transport: torTrasnport,
	}	

	//resp2 := SendGetReq(client,"https://www.whatismyip.com/");
	
	
	
	//fmt.Printf("Body: %s",resp2);

	
	//os.Exit(1);

	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// run task list
	var res string
		chromedp.Navigate(`https://golang.org/pkg/time/`)
		node, err := dom.GetDocument().Do(ctx, h)
		str2, err := dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx, h)
		fmt.Println(str2);
	
	if err != nil {
		fmt.Println(err);
		os.Exit(1);
	}

	fmt.Println(res);
	
	
	//resp1 := SendGetReq(client,"https://recaptcha-demo.appspot.com/recaptcha-v3-request-scores.php");

	//fmt.Println(resp1);
	
	os.Exit(1);
	
	// check if config exists and make a dummy if it does not 
	
	if(!fileExists("config.cfg")) {
		fmt.Printf("Config file not present creating config.cfg!");
		WriteToFile("config.cfg",config_dummy);
		fmt.Printf("Please open up config.cfg and configure it!");	
		return;
	}
	
	
	// -------------------------
	// parse config and login
	
	// read out our config file 
	// and pull the user name / pass 
	// + sessid and how we want to auth
	fp, err := os.OpenFile("config.cfg", os.O_RDONLY, os.ModePerm);
	if(err != nil) {
		fmt.Printf("Failed to open file: %v",err);
		return;
	}

	// close the file when it is no longer needed 
	// automatically
	defer fp.Close();

	sc := bufio.NewScanner(fp);
	
	// read each line in and parse
	// the first section before equals
	// and dump it into a map
	// then use this information to decide
	// how we will authenticate to the server
	
	m := make(map[string]string);
	
	// read every line in split at the
	// equals and load it into a map
	line := 0 // line counter
	for sc.Scan() {
		configstr := sc.Text(); // get the string
		if(configstr == "") { // ignore empty lines
			line++;
			continue;
		}
		fields := strings.Split(configstr,"="); // delimit by eqauls 
		if(len(fields) != 2) { // verify that our input string has been delimited correctly
			fmt.Printf("Error malformed field: '%s' on line %d\n",configstr,line);
			os.Exit(1);
		}
								// strip everything after the space
								// to allow for comments later
		m[fields[0]] = fields[1]; // store into the map
		line++;
	}
	
	
	
	
	// test our map and make sure the user has configured it
	// if they havent quit with a message informing them what to do
	// probably a more compact way to do this but this way allows
	// for more descriptive messages
	if(m["use_session_id"] == "changeme"|| m["use_session_id"] == "") {
		fmt.Println("bot has not been configured for <use_session_id> please read the readme file and configure the bot\n");
		os.Exit(1);
	}
	
	if(m["password"] == "changeme" || m["password"] == "") {
		fmt.Println("bot has not been configured for <password> please read the readme file and configure the bot\n");
		os.Exit(1);
	}
	
	if(m["username"] == "changeme" || m["username"] == "") {
		fmt.Println("bot has not been configured for <username> please read the readme file and configure the bot\n");
		os.Exit(1);
	}
	
	if(m["sessid"] == "changeme" || m["sessid"] == "") {
		fmt.Println("bot has not been configured for <sessid> please read the readme file and configure the bot\n");
		os.Exit(1);
	}
	
	
	if(m["user_agent"] == "changeme" || m["user_agent"] == "") {
		fmt.Println("bot has not been configured for <user_agent> please read the readme file and configure the bot\n");
		os.Exit(1);
	}

	if(m["do_auto"] == "changeme" || m["do_auto"] == "") {
		fmt.Println("bot has not been configured for <do_auto> please read the readme file and configure the bot\n");
		os.Exit(1);
	}

	if(m["do_bootleg"] == "changeme" || m["do_bootleg"] == "") {
		fmt.Println("bot has not been configured for <do_bootleg> please read the readme file and configure the bot\n");
		os.Exit(1);
	}
	
	if(m["do_rackets"] == "changeme" || m["do_rackets"] == "") {
		fmt.Println("bot has not been configured for <do_rackets> please read the readme file and configure the bot\n");
		os.Exit(1);
	}
	
	if(m["ship_state"] == "changeme" || m["ship_state"] == "") {
		fmt.Println("bot has not been configured for <ship_auto> please read the readme file and configure the bot\n");
		os.Exit(1);	
	}
	
	if(m["api_key"] == "changeme" || m["api_key"] == "") {
		fmt.Println("bot has not been configured for <api_key> please read the readme file and configure the bot\n");
		os.Exit(1);			
	}
	
	if(m["user_cookie_val"] == "changeme" || m["user_cookie_val"] == "") {
		fmt.Println("bot has not been configured for <user_cookie_val> please read the readme file and configure the bot\n");
		os.Exit(1);			
	}
	
	
	user_agent = m["user_agent"]; // set the user_agent
	

	
	
	// define our config
	config := &Config{false,false,false,[5]int{0,0,0,0,0},0,0,0,0,0,""};
	config.crime_early_time = 30 + rand.Intn(30);
	if(config.auto_enable) {
		config.aut_early_time = 30 + rand.Intn(30);
	} else {
		config.aut_early_time = 0xffffffff;
	}
	
	if(config.racket_enable) {
		config.rackets_early_time = 30 + rand.Intn(30);
	} else {
		config.rackets_early_time = 0xffffffff;
	}
	
	if(m["do_auto"] == "true") {
		config.auto_enable = true;
	} else if(m["do_auto"] == "false") {
		// nothing its default
	} else {
		fmt.Println("Unknown option for do_auto");
		os.Exit(1);
	}

	config.api_key = m["api_key"];
	if(m["do_bootleg"] == "true") {
		config.bootleg_enable = true;
	} else if(m["do_bootleg"] == "false") {
		// nothing its default
	} else {
		fmt.Println("Unknown option for do_bootleg");
		os.Exit(1);
	}	
	
	if(m["do_rackets"] == "true") {
		config.racket_enable = true;
	} else if(m["do_rackets"] == "false") {
		// nothing its default
	} else {
		fmt.Println("Unknown option for do_rackets");
		os.Exit(1);
	}	
	
	// ship to a random state state
	if(m["ship_state"] == "random") {
		config.ship_state = STATE_RAND;
	} else { // ship to a specific one (falls back on random if same state)
		config.ship_state = GetStateId(m["ship_state"]);
	}
	
	
	// auth with session id 
	if(m["use_session_id"] == "true") {
	
		// not sure if we should be setting cookie expirys
	
		// set a cookie for logging in
		var cookies []*http.Cookie;
		
		loginCookie := &http.Cookie{
			Name: "PHPSESSID",
			Value: m["sessid"],
			Path: "/",
			Domain: "bootleggers.us",
		}
		
		user_cookie_str := fmt.Sprintf("username[%s]",m["username"]);
		
		// the cached login cookie
		userCookie := &http.Cookie{
			Name: user_cookie_str,
			Value: m["user_cookie_val"],
			Path: "/",
			Domain: "bootleggers.us",
		}		
		
		cookies = append(cookies,loginCookie,userCookie);
		cookieURL, _ := url.Parse("https://www.bootleggers.us")
		client.Jar.SetCookies(cookieURL,cookies);
		
		// should send a get to news.php or whatever and verify its logged in
		// before we blast data at the server
		
	} else if(m["use_session_id"] == "false") { 
		
		// or login by using checkuser.php
		// should use resp to check it worked
		SendPostReq(client,"https://www.bootleggers.us/checkuser.php", url.Values{
			"username": {m["username"]},
			"password": {m["password"]},
		})

		if err != nil {
			fmt.Printf("Failed to login");
			os.Exit(1);		
		}		
	} else { // unknown option
		fmt.Printf("Unknown config option for use_session_id: %s(true / false are the expected parameters)\n",m["use_session_id"]);
		os.Exit(1);
	}


	fmt.Println("Logged in!");

	// seed our rng
	rand.Seed(time.Now().UTC().UnixNano())
	
	// sleep off the login
	bot_sleep(4,4,config); 
	
	rand_page :=  "https://www.bootleggers.us/" + rand_pages[rand.Intn(len(rand_pages))];
	
	fmt.Printf("Pulling intial timers from: %s\n",rand_page);
	
	// pull timer bar for the initial timers (should be a random page ideally )
	resp := SendGetReq(client,rand_page);

	doc, err := htmlquery.Parse(strings.NewReader(string(resp)));
	if(err != nil) {
		fmt.Println("Failed to parse the html for xpath");
		os.Exit(1);
	}
	
	//fmt.Println(resp);
	
	//os.Exit(1);
	
	// update all timers
	UpdateTimers(doc,config);
	config.timers[TIMER_RAND] =  296 + rand.Intn(420);

	

	// --------------------------------
	// main bot loop

	// pull our timers get the min value
	// if zero perform the action
	// else sleep for the min time repeat 
	// we will have a function that returns an array of times for this


	// xpath := htmlquery.FindOne(doc, "//input[@type='hidden'][@name='i']");
	// i := htmlquery.SelectAttr(xpath,"value")
	
	
	// when we gui this add a term condition
	
	// work on logging all requests and repsonses
	// with time (params and response codes)
	log("Hello there!");
	
	
	// init our array of function pointers
	var actions[4](func(*http.Client, *Config) bool);
	actions[0] = DoCrime;
	actions[1] = DoAuto;
	actions[2] = DoRackets;
	actions[3] = DoBootleg;
	
	
	var min_timer int = 2147483647; // does go define INT MAX?
	sleep_timer := 0
	
	// 5 - 15 actions sleep for 2-5 mins
	extended_sleep := rand.Intn(5) + 3;
	
	fmt.Println("Bot starting up!");
	
	// sleep off the initial timer pull
	bot_sleep(4,4,config);
	
	for {
		// seed our rng
		rand.Seed(time.Now().UTC().UnixNano())
	
		// Shuffle the functions and perform them do random delays in between
		rand.Shuffle(len(actions), func(i,j int) {
			actions[i], actions[j] = actions[j], actions[i]
		})

		// sleep const + rand delay here
		// and inbetween functions
		

		// call all our funcs (they will check if they are available) 
		for i := 0; i < len(actions); i++ {	
			// if the action was commited sleep a random delay
			if(actions[i](client, config)) {
				// this is what affects actions hitting on time the most
				// adjust if more "on the dot" actions are needed but will raise
				// chances of being caught (more actions has also made this sleep longer)					
				// delay betwen actions
				bot_sleep(4,20,config); 	
			}
		}
		
		// if random timer has elapsed do a random action
		if(config.timers[TIMER_RAND] < 0) {
			fmt.Println("Doing rand_action_sleep!");
			// sleep before doing a random action
			bot_sleep(4,4,config);
			DoRandom(client,config);
		}
		
		min_timer = GetMinTimer(config);
		
		
		// occasionaly visit the page just before the timer is up
		// should make this behavior more complext to randomly peek it 
		// at various times	
		// maybye should also randomize the rate at which it chooses to randomly do it
		// ie rand the 8


		// was originally in each respective do function but i think it works better here
		// as often when a sleep happens the function will totally elapse and this will not 
		// often trigger (this needs heavy testing i dont think its really working...)
		if(rand.Intn(6) == 1 && config.timers[TIMER_CRI] < config.crime_early_time && config.timers[TIMER_CRI] > 0) {
			bot_sleep(4,4,config);
			config.rackets_early_time = 30 + rand.Intn(30);
			fmt.Println("Randomly visiting crime page early!");
			SendGetReq(client,"https://bootleggers.us/crimes.php");	
		} else if(rand.Intn(6) == 1 && config.timers[TIMER_RAC] < config.rackets_early_time && config.timers[TIMER_RAC] > 0) {
			bot_sleep(4,4,config);
			config.rackets_early_time = 30 + rand.Intn(30);
			fmt.Println("Randomly visiting rackets page early!");
			SendGetReq(client,"https://www.bootleggers.us/rackets.php");	
		} else if(rand.Intn(6) == 1 && config.timers[TIMER_AUT] < config.aut_early_time && config.timers[TIMER_AUT] > 0) {
			bot_sleep(4,4,config);
			config.aut_early_time = 30 + rand.Intn(30);
			fmt.Println("Randomly visiting auto page early!");
			SendGetReq(client,"https://www.bootleggers.us/autoburglary.php");	
		}

		

		// seed our rng
		rand.Seed(time.Now().UTC().UnixNano())

		// perform a longer break
		if(config.no_actions >= extended_sleep) {
			config.no_actions = 0; // reset actions
			extended_sleep = rand.Intn(9) + 3; // select number of actions
			sleep_timer = min_timer + rand.Intn(183); 
			fmt.Println("Extended sleep");
		} else { // normal sleep
			sleep_timer = min_timer + rand.Intn(20); 
		}
		
		fmt.Printf("Nothing to do sleeping for %d(%d)...\n",sleep_timer,min_timer);
		
		// this can actually go negative
		if sleep_timer < 2 {
			bot_sleep(4,4,config); // just do a slight action delay
		} else { // sleep the remaining delay off
			bot_sleep(sleep_timer,1,config);
		}
	}
}