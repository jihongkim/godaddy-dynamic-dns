package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "strings"
)

type Configs struct {
    Domain string `json:domain`
    Host string `json:host`
    GoDaddy struct {
        Key string `json:key`
        URL string `json:url`
        Secret string `json:secret`
    } `json:godaddy`
    IPInfo struct {
        Key string `json:key`
        URL string `json:url`
    } `json:ipinfo`
    MyIP string `json:myip`
}

func main() {
    configs, err := loadConfigs()
    if err != nil {
        handleError(err)
        return
    }
    
    hasIPChanged, err := hasIPChanged(configs)

    if !hasIPChanged {
        return
    }

    updatedDNS, err := updateDNS(configs)

    if !updatedDNS {
        return
    }

    fmt.Println("=", configs.MyIP, "=")
    fmt.Println("=", configs.GoDaddy.Key, "=")
    fmt.Println("=", configs.GoDaddy.Secret, "=")

}

func hasIPChanged(configs Configs) (bool, error) {
    response, err := http.Get(configs.IPInfo.URL + "/ip?token=" + configs.IPInfo.Key)
    defer response.Body.Close()
    if err != nil {
        handleError(err)
        return false, errors.New("Could not connect to ipinfo.io")
    }

    bodyBytes, err := ioutil.ReadAll(response.Body)
    if err != nil {
        handleError(err)
        return false, errors.New("Weird body returned from ipinfo.io")
    }

    myip := strings.TrimSpace(string(bodyBytes))
    if myip != configs.MyIP {
        return true, nil
    }

    fmt.Println("Nothing to update")
    return false, nil
}

func updateDNS(configs Configs) (bool, error) {
    godaddyURL := configs.GoDaddy.URL + "/v1/domains/" + configs.Domain + "/records/A/" + configs.Host
    fmt.Println(godaddyURL)
    req, err := http.NewRequest("GET", godaddyURL, nil)
    if err != nil {
        fmt.Println("111111")
        handleError(err)
        return false, errors.New("Could not connect to godaddy.com")
    }
    req.Header.Set("Accept", "application/json")
    req.Header.Set("Authorization", "sso-key "+configs.GoDaddy.Key+":"+configs.GoDaddy.Secret)

    response, err := http.DefaultClient.Do(req)
    if err != nil {
        fmt.Println("22222")
        handleError(err)
        return false, errors.New("Could not connect to godaddy.com")
    }
    defer response.Body.Close()

    bodyBytes, err := ioutil.ReadAll(response.Body)
    if err != nil {
        handleError(err)
        return false, errors.New("Weird body returned from godaddy.com")
    }

    body := strings.TrimSpace(string(bodyBytes))
    fmt.Println("===== body ====")
    fmt.Println("=", body, "=")
    fmt.Println("===============")

    fmt.Println(configs.Domain)
    fmt.Println(configs.Host)

    fmt.Println("Nothing to update")
    return false, nil
}

func handleError(err error) {
    fmt.Println(err)
}

func loadConfigs() (Configs, error) {
    var configs Configs

    configsFile, err := os.Open("configs.json")
    if err != nil {
        return configs, errors.New("Please verify that configs.json file exists")
    }

    jsonParser := json.NewDecoder(configsFile)
    if err = jsonParser.Decode(&configs); err != nil {
        return configs, errors.New("Please verify that configs.json is a valid json file")
    }

    return configs, nil
}
