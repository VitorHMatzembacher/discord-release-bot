package zipbuilder

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

func DownloadAndZip(repos []string, urls []string, output string) error {
	if len(repos) != len(urls) {
		return fmt.Errorf("número de repositórios e URLs não coincide")
	}

	tempDir := os.TempDir()
	finalZipPath := filepath.Join(tempDir, output)

	outFile, err := os.Create(finalZipPath)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo final: %w", err)
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	var wg sync.WaitGroup
	errChan := make(chan error, len(repos))
	var mu sync.Mutex

	for i, url := range urls {
		repo := repos[i]

		wg.Add(1)
		go func(repo, zipURL string) {
			defer wg.Done()
			fmt.Printf("⬇Baixando %s...\n", repo)

			client := &http.Client{}
			req, err := http.NewRequest("GET", zipURL, nil)
			if err != nil {
				errChan <- fmt.Errorf("erro ao criar requisição para %s: %w", repo, err)
				return
			}

			token := os.Getenv("GITHUB_TOKEN")
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			req.Header.Set("Accept", "application/vnd.github+json")
			req.Header.Set("User-Agent", "discord-release-bot")

			resp, err := client.Do(req)
			if err != nil {
				errChan <- fmt.Errorf("erro ao baixar %s: %w", repo, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				errChan <- fmt.Errorf("falha ao baixar %s: %s", repo, string(body))
				return
			}

			fileName := fmt.Sprintf("%s.zip", filepath.Base(repo))

			mu.Lock()
			f, err := zipWriter.Create(fileName)
			if err != nil {
				mu.Unlock()
				errChan <- fmt.Errorf("erro ao criar arquivo no zip final: %w", err)
				return
			}

			_, err = io.Copy(f, resp.Body)
			mu.Unlock()
			if err != nil {
				errChan <- fmt.Errorf("erro ao copiar dados de %s: %w", repo, err)
				return
			}

			fmt.Printf("%s adicionado ao zip final\n", repo)
		}(repo, url)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	fmt.Printf("Arquivo final criado: %s\n", finalZipPath)
	return nil
}
