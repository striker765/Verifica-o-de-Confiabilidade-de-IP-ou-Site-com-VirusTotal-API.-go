package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func checkIPVirusTotal(ip string, apiKey string) (string, error) {
	url := fmt.Sprintf("https://www.virustotal.com/api/v3/ip_addresses/%s", ip)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("x-apikey", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func main() {
	var ip string
	fmt.Print("Digite o IP que deseja verificar: ")
	fmt.Scanln(&ip)

	var apiKey string
	fmt.Print("Digite sua chave de API do VirusTotal: ")
	fmt.Scanln(&apiKey)

	result, err := checkIPVirusTotal(ip, apiKey)
	if err != nil {
		fmt.Println("Erro ao verificar o IP:", err)
		return
	}

	fmt.Println("Resultado:", result)
}
