package register

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/whoisnian/aMonitor-agent/internal/util"
)

const machineIDFile = "/etc/machine-id"

// FetchToken 请求storage取得token
func FetchToken(addr string) string {
	fi, err := os.Open(machineIDFile)
	if err != nil {
		log.Panicln(err)
	}
	defer fi.Close()

	var req struct{ MachineID string }
	req.MachineID = strings.TrimSpace(string(util.ReadAll(fi)))

	content, _ := json.Marshal(req)
	url := "http://" + addr + "/register"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(content))
	if err != nil {
		log.Panicln(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Panicln(url + " return " + resp.Status)
	}

	var data struct{ Token string }
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Panicln(err)
	}

	return data.Token
}
