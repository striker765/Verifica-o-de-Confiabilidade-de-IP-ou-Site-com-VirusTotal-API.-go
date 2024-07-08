package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Configurações do MySQL
const (
	mysqlUsername = "$MYSQL_USERNAME"
	mysqlPassword = "$MYSQL_PASSWORD"
	mysqlHost     = "endereco_do_seu_servidor_mysql"
	mysqlDatabase = "nome_do_seu_banco_de_dados"
)

// Configurações do GitLab
const (
	gitlabURL     = "https://gitlab.com"
	gitlabUserID  = "seu_user_id"       // Seu ID de usuário do GitLab
	gitlabToken   = "glpat-kHJXvuTqY2sUNRBQ3a_Z" // Seu token de acesso do GitLab
	gitlabProject = "seu_project_id"   // ID do seu projeto no GitLab
)

// Consulta SQL para encontrar lojas indisponíveis nos últimos 30 dias
const sqlQuery = `
SELECT 
    h.name AS loja, 
    FROM_UNIXTIME(e.clock) AS inicio_indisponibilidade, 
    FROM_UNIXTIME(e.clock + e.value) AS fim_indisponibilidade, 
    (e.clock + e.value) - e.clock AS duracao_segundos, 
    SEC_TO_TIME((e.clock + e.value) - e.clock) AS duracao_formatada 
FROM 
    events e 
JOIN 
    hosts h ON e.objectid = h.hostid 
WHERE 
    e.source = '0' 
    AND e.object = '0' 
    AND e.value = '1' 
    AND h.name LIKE 'FastShop%' 
    AND e.clock > UNIX_TIMESTAMP(DATE_SUB(NOW(), INTERVAL 30 DAY)) 
ORDER BY 
    h.name, inicio_indisponibilidade;
`

// Função para conectar ao MySQL e executar a consulta SQL
func executeSQLQuery() (*sql.Rows, error) {
	// Monta a string de conexão
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", mysqlUsername, mysqlPassword, mysqlHost, mysqlDatabase)

	// Conecta ao MySQL
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao MySQL: %v", err)
	}
	defer db.Close()

	// Executa a consulta SQL
	rows, err := db.Query(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("erro ao executar consulta SQL: %v", err)
	}

	return rows, nil
}

// Função para formatar os dados em CSV
func formatToCSV(rows *sql.Rows) ([]string, error) {
	defer rows.Close()

	// Recupera os nomes das colunas
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("erro ao recuperar nomes das colunas: %v", err)
	}

	// Cria o slice para armazenar as linhas do CSV
	var csvData []string

	// Adiciona o cabeçalho CSV
	csvData = append(csvData, strings.Join(columns, ","))

	// Itera pelas linhas do resultado
	for rows.Next() {
		// Cria um slice para armazenar os valores das colunas
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Escaneia os valores das colunas na linha atual
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("erro ao escanear valores das colunas: %v", err)
		}

		// Converte os valores em strings
		var valueStrings []string
		for _, value := range values {
			switch v := value.(type) {
			case nil:
				valueStrings = append(valueStrings, "NULL")
			case []byte:
				valueStrings = append(valueStrings, string(v))
			default:
				valueStrings = append(valueStrings, fmt.Sprintf("%v", v))
			}
		}

		// Adiciona a linha formatada ao CSV
		csvData = append(csvData, strings.Join(valueStrings, ","))
	}

	return csvData, nil
}

// Função para enviar o arquivo CSV para o GitLab
func uploadToGitLab(csvData []string) error {
	// Formata o conteúdo CSV como texto
	csvContent := strings.Join(csvData, "\n")

	// Configura os cabeçalhos HTTP
	headers := map[string]string{
		"PRIVATE-TOKEN": gitlabToken,
		"Content-Type":  "text/plain",
	}

	// Configura os parâmetros da requisição
	params := map[string]string{
		"branch":         "main",
		"commit_message": "Adicionando resultados da consulta SQL",
	}

	// Monta a URL para upload no GitLab
	url := fmt.Sprintf("%s/api/v4/projects/%s/repository/files/results.csv", gitlabURL, gitlabProject)

	// Cria uma requisição POST para enviar o arquivo CSV para o GitLab
	resp, err := requests.Post(url, headers, params, csvContent)
	if err != nil {
		return fmt.Errorf("erro ao enviar arquivo para o GitLab: %v", err)
	}
	defer resp.Body.Close()

	// Verifica o código de status da resposta
	if resp.StatusCode != 201 {
		return fmt.Errorf("falha ao enviar arquivo para o GitLab. Status Code: %d", resp.StatusCode)
	}

	fmt.Println("Arquivo 'results.csv' enviado com sucesso para o GitLab.")
	return nil
}

// Função principal para executar o script
func main() {
	// Executa a consulta SQL
	rows, err := executeSQLQuery()
	if err != nil {
		log.Fatalf("Erro ao executar consulta SQL: %v", err)
	}

	// Formata os dados em CSV
	csvData, err := formatToCSV(rows)
	if err != nil {
		log.Fatalf("Erro ao formatar dados em CSV: %v", err)
	}

	// Envia o arquivo CSV para o GitLab
	if err := uploadToGitLab(csvData); err != nil {
		log.Fatalf("Erro ao enviar arquivo para o GitLab: %v", err)
	}
}
