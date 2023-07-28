package main

import (
  base64 "encoding/base64"
  "bytes"
  "encoding/json"
  "errors"
  "net/http"
  "time"
  "fmt"
  "log"
  "syscall"
  "unsafe"
  "bufio"
	"os"
  "os/exec"
)

type srb struct { // srb = slack request body
    Text string `json:"text"`
}

//------------------------------------------------------------------------------
// Slack webhook function
//------------------------------------------------------------------------------
func ssn(webhookUrl string, msg string) error { // ssn = SendSlackNotification

    decoded_url, err := base64.StdEncoding.DecodeString(webhookUrl)
    if err != nil { panic(err); } // this error handling sucks

    slackBody, _ := json.Marshal(srb{Text: msg})
    req, err := http.NewRequest(http.MethodPost, string(decoded_url), bytes.NewBuffer(slackBody))
    if err != nil { return err }

    req.Header.Add("Content-Type", "application/json")
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil { return err }

    buf := new(bytes.Buffer)
    buf.ReadFrom(resp.Body)
    if buf.String() != "ok" { return errors.New("Non-ok response returned") }
    return nil
}

//------------------------------------------------------------------------------
// Main
//------------------------------------------------------------------------------
func main() {
  fmt.Println("============");
  fmt.Println("= CT Agent =");
  fmt.Println("============\n");

  // Base64 encoded slack webhook url
  webhookUrl := "base64 encoded webook url here"

  // Attempt to open the config file first with listing of all folders / files
  // and if successful load config into scanner and read all lines into variable
  // for use

  file, err := os.Open("/etc/ct.conf");
  if err != nil { log.Fatalf("failed opening file: %s", err); os.Exit(1); }
  fmt.Println("Config loaded");

  var config []string
  scanner := bufio.NewScanner(file);
	scanner.Split(bufio.ScanLines);
	for scanner.Scan() { config = append(config, scanner.Text()) }
	file.Close()

  // Setup Inotify
  fmt.Println("Initializing...")
  fd, err := syscall.InotifyInit()
  if err != nil { log.Fatal(err); }
  defer syscall.Close(fd)

/*
IN_ACCESS_ALL = File was read/written or executed (anything)
IN_ACCESS = File was accessed (read/exec)
IN_MODIFY = File was modified (write/truncate)
IN_OPEN   = File was opened
*/

  // build list of watchers from data we read in from our config file.
  wd, err := syscall.InotifyAddWatch(fd, config[0], syscall.IN_ACCESS) // Built first entry and define wd var
  if err != nil { log.Fatal(err); }
  for i:=1; i< len(config); i++ { _, err = syscall.InotifyAddWatch(fd, config[i], syscall.IN_ACCESS); } // Now build all remaining
  defer syscall.InotifyRmWatch(fd, uint32(wd)); // Setup to get rid of all watchers if process killed/terminated

  fmt.Println("Ready!")

//------------------------------------------------------------------------------
// Main logic loop
//------------------------------------------------------------------------------
  for {
      buffer := make([]byte, syscall.SizeofInotifyEvent*128) // Room for 128 Events
      bytesRead, err := syscall.Read(fd, buffer);
      if err != nil { log.Fatal(err); }
      if bytesRead < syscall.SizeofInotifyEvent { /* No point trying if we don't have at least one event */ continue }
// -----------------------------------------------------------------------------
//     		fmt.Printf("Size of InotifyEvent is %s\n", syscall.SizeofInotifyEvent)
//     		fmt.Printf("Bytes read: %d\n", bytesRead)

      		offset := 0 // start at begining
      		for offset < bytesRead-syscall.SizeofInotifyEvent {
      			event := (*syscall.InotifyEvent)(unsafe.Pointer(&buffer[offset]))
      			fmt.Printf("%+v\n", event)
//     			if (event.Mask & syscall.IN_ACCESS) > 0 { fmt.Printf("Saw IN_ACCESS for %+v\n", event) }
//     			if (event.Mask & syscall.IN_MODIFY) > 0 { fmt.Printf("Saw IN_MODIFY for %+v\n", event) }
//     			if (event.Mask & syscall.IN_OPEN) > 0 { fmt.Printf("Saw IN_OPEN for %+v\n", event) }
      			// We need to account for the length of the name
      			offset += syscall.SizeofInotifyEvent + int(event.Len)
          }
//-------------------------------------------------------------------------------------

      cmd := "lsof -i -P -n | head -n 20";       // Get current connections (first 20)
      connections, _ := exec.Command("bash", "-c", cmd).Output();

      cmd = "last -5"; // Get uptime, users, on box
      last, _ := exec.Command("bash", "-c", cmd).Output();

      cmd = "w";       // see who is on machine and load / uptime
      who, _ := exec.Command("bash", "-c", cmd).Output();

      cmd = "hostname";
      hostname, _ := exec.Command("bash", "-c", cmd).Output();

      cmd = "hostname -I";       // get ip address of machine
      ip_addr, _ := exec.Command("bash", "-c", cmd).Output();

      event := (*syscall.InotifyEvent)(unsafe.Pointer(&buffer[0]))

      notification := "*Alert Information*" + "\n" +

                      "```" + "Hostname: " + string(hostname) +
                      "System IP: " + string(ip_addr) +

                      "Accessed File/Path: " + config[(event.Wd)-1] + "\n" + "```" +

                      "\n" +

                      "*Who is logged in*" + "\n" +
                      "```"+string(who) + "```" + "\n" +

                      "\n" +

                      "*Last 5 logins*" + "\n" +
                      "```" + string(last) + "```" + "\n" +

                      "\n" +

                      "*First 20 current Connections*" + "\n" +
                      "```" + string(connections) + "```" + "\n" +
                      "\n <!channel|channel>" + "\n" +
                      "\n"

      buffer = make([]byte, syscall.SizeofInotifyEvent*128); // clear buffer
      bytesRead = 0;
      offset=0;
      err = ssn(webhookUrl, notification);
      if err != nil { log.Fatal(err); }
      } // End of for loop above

} // End of main
