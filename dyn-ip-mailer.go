package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "net/smtp"
    "github.com/spf13/viper"
)

const IPIFY_URL = "https://api.ipify.org"

type Config struct {
    CachedIpFilename string
    SmtpServer       string
    SmtpPort         int
    SmtpUsername     string
    SmtpPassword     string
    EmailAddress     string
}

func getSettings() (Config, error) {
    var config Config

    viper.SetConfigName(".dyn-ip-mailer")
    viper.AddConfigPath("$HOME")
    viper.AddConfigPath(".")

    if err := viper.ReadInConfig(); err != nil {
        return config, err
    }

    if ! viper.IsSet("cachefile") {
        return config, fmt.Errorf("value cachefile not found in config")
    }
    config.CachedIpFilename = viper.GetString("cachefile")

    if ! viper.IsSet("email") {
        return config, fmt.Errorf("value email not found in config")
    }
    config.EmailAddress = viper.GetString("email")

    if ! viper.IsSet("smtp") {
        return config, fmt.Errorf("section smtp not found in config")
    }
    smtp := viper.Sub("smtp")

    if ! smtp.IsSet("server") {
        return config, fmt.Errorf("value smtp.server not found in config")
    }
    config.SmtpServer = smtp.GetString("server")

    if ! smtp.IsSet("port") {
        return config, fmt.Errorf("value smtp.port not found in config")
    }
    config.SmtpPort = smtp.GetInt("port")

    if ! smtp.IsSet("username") {
        return config, fmt.Errorf("value smtp.username not found in config")
    }
    config.SmtpUsername = smtp.GetString("username")

    if ! smtp.IsSet("password") {
        return config, fmt.Errorf("value smtp.password not found in config")
    }
    config.SmtpPassword = smtp.GetString("password")

    return config, nil
}

func getCurrentIp() (string, error) {
    resp, err := http.Get(IPIFY_URL)
    if err != nil {
        return "", fmt.Errorf("http to IPIFY failed: %v", err)
    }
    defer resp.Body.Close()

    body_bytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("failed reading IPIFY body: %v", err)
    }

    return string(body_bytes), nil
}

// Determine what the IP was last time we checked.
// Returns string 'none' when cache file does not exist.
func getCachedIp(cache_file string) (string, error) {
    var cached_ip = "none"
    if _, err := os.Stat(cache_file); err == nil {
        // Read the cached IP address from previous check
        content, err := ioutil.ReadFile(cache_file)
        if err != nil {
            return "", fmt.Errorf("failed reading cache file: %v", err)
        }
        cached_ip = string(content)
    }
    return cached_ip, nil
}

func updateCachedIp(newip string, cachefile string) {
    ioutil.WriteFile(cachefile, []byte(newip), 0644)
}

func sendEmail(newip string, config Config) {
    auth := smtp.PlainAuth("", config.SmtpUsername, config.SmtpPassword,
        config.SmtpServer)
    msg := []byte(
       "To: " + config.EmailAddress + "\r\n" +
       "Subject: new dynamic IP detected\r\n" +
       "\r\n" +
       "IP: " + newip + "\r\n")
    err := smtp.SendMail(
        fmt.Sprintf("%v:%v", config.SmtpServer, config.SmtpPort),
        auth, config.EmailAddress, []string{config.EmailAddress}, msg)

    if err != nil {
        fmt.Println("Failed sending email: ", err)
    } else {
        fmt.Println("Mail sent")
    }
}

func main() {
    config, err := getSettings()
    if err != nil {
        panic(err)
    }

    current_ip, err := getCurrentIp()
    if err != nil {
        panic(err)
    }
    fmt.Println("Current external IP:", current_ip)

    cached_ip, err := getCachedIp(config.CachedIpFilename)
    if err != nil {
        panic(err)
    }
    fmt.Println(cached_ip)

    if current_ip != cached_ip {
        fmt.Println("IP changed:", cached_ip, "->", current_ip)
        updateCachedIp(current_ip, config.CachedIpFilename)
        sendEmail(current_ip, config)
    } else {
        fmt.Println("Ip unchanged")
    }
}
