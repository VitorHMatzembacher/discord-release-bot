package main

import (
	"encoding/json"
	"fmt"
	"github-release-bot/internal/github"
	zip "github-release-bot/internal/zip"
	"log"
	"os"
	"path/filepath"
)

func main() {
	pkg := os.Getenv("PACKAGE")
	if pkg == "" {
		log.Fatal("Variável PACKAGE não definida.")
	}

	reposFile := filepath.Join("repositories", fmt.Sprintf("%s.json", pkg))
	fileData, err := os.ReadFile(reposFile)
	if err != nil {
		log.Fatalf("Erro ao ler arquivo %s: %v", reposFile, err)
	}

	var repos []string
	if err := json.Unmarshal(fileData, &repos); err != nil {
		log.Fatalf("Erro ao processar JSON: %v", err)
	}

	var urls []string
	for _, repo := range repos {
		tag, err := github.GetLatestTag(repo)
		if err != nil {
			log.Printf("Erro ao buscar tag de %s: %v", repo, err)
			continue
		}
		urls = append(urls, tag.ZipballURL)
	}

	output := "/tmp/latest_releases.zip"
	if err := zip.DownloadAndZip(repos, urls, output); err != nil {
		log.Fatalf("Erro ao gerar ZIP: %v", err)
	}

	log.Printf("Pacote final criado em: %s\n", output)
}
