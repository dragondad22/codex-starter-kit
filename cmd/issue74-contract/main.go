// Command issue74-contract performs the read-only live qualification required by issue #74.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/dragondad22/codex-starter-kit/githubadapter"
)

const (
	repositoryOwner = "codex-starter-kit-labs"
	repositoryName  = "codex-starter-kit-sandbox"
	appID           = "4319725"
	installationID  = "147093185"
	appActor        = "codex-starter-kit-labs-reconciler"
	accountID       = "305967668"
)

var (
	commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
	digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
)

type receipt struct {
	SchemaVersion      int       `json:"schema_version"`
	EvidenceMode       string    `json:"evidence_mode"`
	Outcome            string    `json:"outcome"`
	Actor              string    `json:"actor"`
	Account            string    `json:"account"`
	InstallationID     string    `json:"installation_id"`
	PermissionSource   string    `json:"permission_source"`
	PermissionRevision string    `json:"permission_revision"`
	Permissions        []string  `json:"permissions"`
	Repository         string    `json:"repository"`
	Revision           string    `json:"revision"`
	Path               string    `json:"path"`
	Digest             string    `json:"digest"`
	ObservedAt         time.Time `json:"observed_at"`
}

func main() {
	if err := run(context.Background(), os.Args[1:], os.Getenv, http.DefaultClient, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, getenv func(string) string, client *http.Client, output io.Writer) error {
	flags := flag.NewFlagSet("issue74-contract", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	revision := flags.String("revision", "", "immutable sandbox commit SHA")
	path := flags.String("path", "", "repository-relative file path")
	expectedDigest := flags.String("expected-digest", "", "expected SHA-256 content digest")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return errors.New("revision, path, and expected-digest flags are required; positional arguments are unsupported")
	}
	if !commitPattern.MatchString(*revision) || !safePath(*path) || !digestPattern.MatchString(*expectedDigest) {
		return errors.New("qualification input requires an immutable revision, safe file path, and SHA-256 digest")
	}
	privateKey := getenv("CSK_APP_PRIVATE_KEY")
	if privateKey == "" {
		return errors.New("GitHub App private key is unavailable")
	}
	provider, err := githubadapter.NewAppInstallationProvider(githubadapter.AppInstallationConfig{
		RESTBaseURL: "https://api.github.com", APIVersion: "2026-03-10", AppID: appID,
		InstallationID: installationID, Actor: appActor, Account: repositoryOwner, AccountID: accountID,
	}, githubadapter.PrivateKeyProviderFunc(func(context.Context) ([]byte, error) {
		return []byte(privateKey), nil
	}), client)
	if err != nil {
		return err
	}
	credential, err := provider.Credential(ctx)
	if err != nil {
		return err
	}
	if !slices.Contains(credential.Permissions, "contents:read") {
		return errors.New("GitHub App installation lacks required contents:read permission")
	}

	digest, err := observeContent(ctx, client, credential.Token, *revision, *path)
	if err != nil {
		return err
	}
	if digest != *expectedDigest {
		return errors.New("live repository content does not match the approved digest")
	}
	result := receipt{
		SchemaVersion: 1, EvidenceMode: "live", Outcome: "pass", Actor: credential.Actor,
		Account: credential.Account, InstallationID: credential.InstallationID,
		PermissionSource: credential.PermissionSource, PermissionRevision: credential.PermissionRevision,
		Permissions: slices.Clone(credential.Permissions), Repository: repositoryOwner + "/" + repositoryName,
		Revision: *revision, Path: *path, Digest: digest, ObservedAt: time.Now().UTC(),
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func observeContent(ctx context.Context, client *http.Client, token, revision, path string) (string, error) {
	endpoint := "https://api.github.com/repos/" + repositoryOwner + "/" + repositoryName + "/contents/" + escapePath(path) + "?ref=" + url.QueryEscape(revision)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", errors.New("prepare live content observation")
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("X-GitHub-Api-Version", "2026-03-10")
	request.Header.Set("User-Agent", "codex-starter-kit")
	response, err := client.Do(request)
	if err != nil {
		return "", errors.New("live content observation is offline")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("live content observation was denied with HTTP %d", response.StatusCode)
	}
	var content struct {
		Type     string `json:"type"`
		Encoding string `json:"encoding"`
		Content  string `json:"content"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&content); err != nil || content.Type != "file" || content.Encoding != "base64" {
		return "", errors.New("live content observation returned an invalid file response")
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(content.Content, "\n", ""))
	if err != nil {
		return "", errors.New("live content observation returned invalid base64")
	}
	digest := sha256.Sum256(decoded)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func safePath(path string) bool {
	if path == "" || strings.HasPrefix(path, "/") || strings.Contains(path, `\`) {
		return false
	}
	for _, part := range strings.Split(path, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}

func escapePath(path string) string {
	parts := strings.Split(path, "/")
	for index := range parts {
		parts[index] = url.PathEscape(parts[index])
	}
	return strings.Join(parts, "/")
}
