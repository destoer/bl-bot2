package main

import (
	"os"
	"bufio"
	"fmt"

)

/* reopening the log file might be too slow */
/* but we aernt doing many logs so it might be fine... */
func log(a ...interface{}) {
	file, err := os.OpenFile("log.txt", os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0644);
	if err != nil {
		fmt.Println("Could not open file!");
		fmt.Println(err);
		os.Exit(1);
	}
	
	defer file.Close();
	
	w := bufio.NewWriter(file);
	
	fmt.Fprintln(w,a...);
	
	w.Flush();
}