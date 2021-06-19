package main

import (
	"fmt"
	"github.com/judwhite/go-svc"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var folderPath = getBasePath() // need to get base path in order for win service to know the path

var wg sync.WaitGroup // WaitGroup is used to wait for the program to finish goroutines.

// program implements svc.Service
type program struct {
	LogFile *lumberjack.Logger
	wg      sync.WaitGroup
	quit    chan struct{}
}

func main() {
	prg := &program{}

	// Call svc.Run to start your program/service.
	if err := svc.Run(prg); err != nil {
		log.Fatal(err)
	}

}

func (p *program) Init(env svc.Environment) error {

	logPath := "log-worker.log"
	// Set the Lumberjack logger
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    100, // megabytes
		MaxBackups: 5,
		LocalTime:  true,
		Compress:   true, // disabled by default
	}

	// Multi writer to stderr and file
	mWriter := io.MultiWriter(os.Stderr, lumberjackLogger)
	log.SetOutput(mWriter)

	log.Printf("is win service? %v\n", env.IsWindowsService())
	// write to log when running as a Windows Service
	if env.IsWindowsService() {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Error(err)
			return err
		}
		logPath = filepath.Join(dir, "log-worker.log")
		log.Println("Log path", logPath)
		p.LogFile = lumberjackLogger
		log.SetOutput(lumberjackLogger)
	}

	return nil
}

func (p *program) Start() error {
	// The Start method must not block, or Windows may assume your service failed
	// to start. Launch a Goroutine here to do something interesting/blocking.

	p.quit = make(chan struct{})

	p.wg.Add(1)
	go func() {
		log.Println("Starting...")

		ticker := time.NewTicker(60 * time.Second) // service interval
		quit := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker.C:
					// do service stuff

					wg.Add(4)
					// create initial dir's
					go createInitDir("audit")
					go createInitDir("system")

					// delete old files
					go deleteOldFile(folderPath + "\\audit\\")
					go deleteOldFile(folderPath + "\\system\\")

					exportFile() // export milestone logs to csv

					wg.Wait()

				case <-quit:
					ticker.Stop()
					return
				}
			}
		}()

		<-p.quit
		log.Println("Quit signal received...")
		p.wg.Done()
	}()

	return nil
}

func (p *program) Stop() error {
	// The Stop method is invoked by stopping the Windows service, or by pressing Ctrl+C on the console.
	// This method may block, but it's a good idea to finish quickly or your process may be killed by
	// Windows during a shutdown/reboot. As a general rule you shouldn't rely on graceful shutdown.

	log.Println("Stopping...")
	close(p.quit)
	p.wg.Wait()
	log.Println("Stopped.")
	return nil
}

func getBasePath() string {
	basePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return basePath
}

func createInitDir(dirName string) {
	defer wg.Done() // Schedule the call to WaitGroup's Done to tell goroutine is completed.
	_, err := os.Stat(folderPath + "\\" + dirName)
	if os.IsNotExist(err) {
		log.Println("Creating initial dir...", dirName)
		err := os.Mkdir(folderPath+"\\"+dirName, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func deleteOldFile(filePath string) {
	defer wg.Done()
	var cutoff = 168 * time.Hour // file retention time
	fileInfo, err := ioutil.ReadDir(filePath)
	if err != nil {
		log.Error(err.Error())
	}
	now := time.Now()
	for _, info := range fileInfo {
		if diff := now.Sub(info.ModTime()); diff > cutoff {
			log.Printf("Deleting %s which is %s old\n", info.Name(), diff)
			err := os.Remove(filePath + "\\" + info.Name())
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func exportFile() {
	// get cuurent timestamp
	currTimestamp := time.Now()

	// get current timestamp - n minutes, seconds rounded to 0.
	prevMin1 := currTimestamp.Add(time.Duration(-1) * time.Minute).Format("2006-01-02 15:04:00")
	prevMin2 := currTimestamp.Add(time.Duration(-2) * time.Minute).Format("2006-01-02 15:04:00")

	// replace symbols for valid windows file name
	prevName1 := strings.ReplaceAll(prevMin1, ":", "-")
	prevName2 := strings.ReplaceAll(prevMin2, ":", "-")

	// Run powershell commands
	cmd := exec.Command("powershell", "-nologo", "-noprofile")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer stdin.Close()

		// do both jobs synchronously in single Powershell session to reduce calls to management server
		fmt.Fprintln(stdin, "Connect-ManagementServer -Server localhost -AcceptEula")
		fmt.Fprintln(stdin, `Get-Log -LogType Audit -BeginTime "`+prevMin2+`" -EndTime "`+prevMin1+`" | Foreach-Object { $_.'Message text' = $_.'Message text'.Replace("`+"`r"+"`n"+"\""+`, " "); $_ } `+` | `+`Export-Csv -Path "`+folderPath+"\\audit\\"+prevName2+` `+prevName1+` audit.csv" -NoTypeInformation -Encoding UTF8`)
		fmt.Fprintln(stdin, `Get-Log -LogType System -BeginTime "`+prevMin2+`" -EndTime "`+prevMin1+`" | Foreach-Object { $_.'Message text' = $_.'Message text'.Replace("`+"`r"+"`n"+"\""+`, " "); $_ } `+` | `+`Export-Csv -Path "`+folderPath+"\\system\\"+prevName2+` `+prevName1+` system.csv" -NoTypeInformation -Encoding UTF8`)
		fmt.Fprintln(stdin, "Disconnect-ManagementServer")
	}()
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s\n", out)
}
