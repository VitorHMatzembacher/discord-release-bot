package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Tag struct {
	Name       string `json:"name"`
	ZipballURL string `json:"zipball_url,omitempty"`
}

func GetLatestTag(repo string) (*Tag, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("token do GitHub não encontrado")
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/tags", repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "discord-release-bot")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("falha ao buscar tags (%s): %s", repo, string(body))
	}

	var tags []Tag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}

	if len(tags) == 0 {
		return nil, fmt.Errorf("nenhuma tag encontrada para o repositório %s", repo)
	}

	latest := tags[0]

	zipURL := fmt.Sprintf("https://github.com/%s/archive/refs/tags/%s.zip", repo, latest.Name)

	return &Tag{
		Name:       latest.Name,
		ZipballURL: zipURL,
	}, nil
}
