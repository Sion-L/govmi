package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime/pprof"
	"strings"
)

const (
	JmsServerURL = "http://jms.firecloud.wan"
	UserName     = "lilang"
	Password     = "ll772576"
)

type HostInfo struct {
	Id               string        `json:"id"`
	Hostname         string        `json:"hostname"`
	Ip               string        `json:"ip"`
	Platform         string        `json:"platform"`
	Protocols        []string      `json:"protocols"`
	IsActive         bool          `json:"is_active"`
	PublicIp         interface{}   `json:"public_ip"`
	Comment          string        `json:"comment"`
	Number           interface{}   `json:"number"`
	Vendor           string        `json:"vendor"`
	Model            string        `json:"model"`
	Sn               string        `json:"sn"`
	CpuModel         string        `json:"cpu_model"`
	CpuCount         int           `json:"cpu_count"`
	CpuCores         int           `json:"cpu_cores"`
	CpuVcpus         int           `json:"cpu_vcpus"`
	Memory           string        `json:"memory"`
	DiskTotal        string        `json:"disk_total"`
	DiskInfo         string        `json:"disk_info"`
	Os               string        `json:"os"`
	OsVersion        string        `json:"os_version"`
	OsArch           string        `json:"os_arch"`
	HostnameRaw      string        `json:"hostname_raw"`
	HardwareInfo     string        `json:"hardware_info"`
	Connectivity     string        `json:"connectivity"`
	DateVerified     string        `json:"date_verified"`
	Domain           interface{}   `json:"domain"`
	AdminUser        string        `json:"admin_user"`
	AdminUserDisplay string        `json:"admin_user_display"`
	Nodes            []string      `json:"nodes"`
	NodesDisplay     []string      `json:"nodes_display"`
	Labels           []interface{} `json:"labels"`
	CreatedBy        string        `json:"created_by"`
	DateCreated      string        `json:"date_created"`
	OrgId            string        `json:"org_id"`
	OrgName          string        `json:"org_name"`
}

type name struct {
}

type CreateHosts struct {
	Id           string   `json:"id"`
	Hostname     string   `json:"hostname"`
	Ip           string   `json:"ip"`
	Platform     string   `json:"platform"`
	Protocols    []string `json:"protocols"`
	Protocol     string   `json:"protocol"`
	Port         int      `json:"port"`
	IsActive     bool     `json:"is_active"`
	PublicIp     string   `json:"public_ip"`
	Comment      string   `json:"comment"`
	Domain       string   `json:"domain"`
	AdminUser    string   `json:"admin_user"`
	Nodes        []string `json:"nodes"`
	NodesDisplay []string `json:"nodes_display"`
	Labels       []string `json:"labels"`
}

func GetToken(jmsurl, username, password string) (string, error) {
	url := jmsurl + "/api/v1/authentication/auth/"
	query_args := strings.NewReader(`{
        "username": "` + username + `",
        "password": "` + password + `"
    }`)
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, query_args)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	debug / pprof.Profile{}
	response := map[string]interface{}{}
	json.Unmarshal(body, &response)
	return response["token"].(string), nil
}

func GetUserInfo(jmsurl, token string) []byte {
	url := jmsurl + "/api/v1/assets/assets/"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("X-JMS-ORG", "00000000-0000-0000-0000-000000000002")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return body
}

func main() {
	var hostinfo []*HostInfo
	token, err := GetToken(JmsServerURL, UserName, Password)
	if err != nil {
		log.Fatal(err)
	}

	body := GetUserInfo(JmsServerURL, token)
	err = json.Unmarshal(body, &hostinfo)
	if err != nil {
		fmt.Println(err)
	}

	for _, v := range hostinfo {
		fmt.Println(v.NodesDisplay)
	}

}
