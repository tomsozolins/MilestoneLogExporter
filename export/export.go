package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup
var mgmtSrv string
var logType string

const MAX = 16

func main() {
	mgmtSrv = os.Args[1]
	logType = os.Args[2]
	begin_timestamp := os.Args[3]
	end_timestamp := os.Args[4]

	start, err := time.Parse("2006-01-02 15:04:00", begin_timestamp)
	if err != nil {
		log.Println(err)

	}

	end, err := time.Parse("2006-01-02 15:04:00", end_timestamp)
	if err != nil {
		log.Println(err)
	}

	kvs := map[string]string{}

	for d := start; !d.After(end); d = d.Add(time.Duration(+24) * time.Hour) {
		// build key value map
		kvs[d.Format("2006-01-02 15:04:00")] = d.Add(time.Duration(+24) * time.Hour).Format("2006-01-02 15:04:00")
	}

	lenKvs := len(kvs)
	wg.Add(lenKvs)

	// https://stackoverflow.com/a/25306439
	sem := make(chan int, MAX)
	for k, v := range kvs {
		sem <- 1 // will block if there is MAX ints in sem
		log.Println(k, v)
		go exportFile(k, v, sem)
		// go exportFile(k, v)
	}

	wg.Wait()

}

// func exportFile(begin, end string)
func exportFile(begin, end string, sem chan int) {
	defer wg.Done()

	n1 := strings.ReplaceAll(begin, ":", "-")
	n2 := strings.ReplaceAll(end, ":", "-")
	logTypeName := strings.ToLower(logType)

	cmd := exec.Command("powershell", "-nologo", "-noprofile")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Exporting logs...")

	go func() {
		defer stdin.Close()

		fmt.Fprintln(stdin, "Connect-ManagementServer -Server "+mgmtSrv+" -AcceptEula")
		fmt.Fprintln(stdin, `Get-Log -LogType `+logType+` -BeginTime "`+begin+`" -EndTime "`+end+`" | Foreach-Object { $_.'Message text' = $_.'Message text'.Replace("`+"`r"+"`n"+"\""+`, " "); $_ } `+`| `+`Export-Csv -Path ".\\`+n1+" "+n2+" "+logTypeName+`.csv" -NoTypeInformation -Encoding UTF8`)
		fmt.Fprintln(stdin, "Disconnect-ManagementServer")
	}()
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s\n", out)

	<-sem // removes an int from sem, allowing another to proceed
}
