// LLM inference in Helix.ml from Dagger
//
// This module lets you call LLM inference on open source models in Helix.ml
//
// For more info, see https://helixml.substack.com/p/introducing-the-helix-cli-and-dagger

package main

import (
	"context"
	"strings"
)

type Helix struct{}

// example usage: "dagger call get-secret --helix-credentials ~/.helix/credentials"
func (m *Helix) GetSecret(ctx context.Context, helixCredentials *File) (string, error) {
	ctr, err := m.WithHelixSecret(ctx, dag.Container().From("ubuntu:latest"), helixCredentials)
	if err != nil {
		return "", err
	}
	return ctr.
		WithExec([]string{"bash", "-c", "cat /root/.helix/credentials |base64"}).
		Stdout(ctx)
}

func (m *Helix) WithHelixSecret(ctx context.Context, ctr *Container, helixCredentials *File) (*Container, error) {
	credsFile, err := helixCredentials.Contents(ctx)
	if err != nil {
		return nil, err
	}
	/* we expect ~/.helix/credentials to be of the form:
	HELIX_API_KEY=a1b2
	HELIX_API_URL=http://localhost
	*/
	var apiKey string
	var apiUrl string = "https://app.tryhelix.ai"
	var activeTools string
	lines := strings.Split(credsFile, "\n")
	for _, line := range lines {
		if len(line) > 0 && string(line[0]) == "#" {
			continue
		}
		if strings.Contains(line, "HELIX_API_KEY=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) > 1 {
				apiKey = parts[1]
			}
		}
		if strings.Contains(line, "HELIX_API_URL=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) > 1 {
				apiUrl = parts[1]
			}
		}
		if strings.Contains(line, "HELIX_ACTIVE_TOOLS=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) > 1 {
				activeTools = parts[1]
			}
		}
	}
	secret := dag.SetSecret("helix-api-key", apiKey)
	ctr = ctr.
		WithSecretVariable("HELIX_API_KEY", secret).
		WithEnvVariable("HELIX_API_URL", apiUrl).
		WithEnvVariable("HELIX_ACTIVE_TOOLS", activeTools)
	return ctr, nil
}

func (m *Helix) HelixCli(ctx context.Context, helixCredentials *File) (*Container, error) {
	ctr := dag.Container().
		// From("helix-controlplane:dev")
		From("europe-docker.pkg.dev/helixml/helix/controlplane-dev:dev002")
	ctr, err := m.WithHelixSecret(ctx, ctr, helixCredentials)
	if err != nil {
		return nil, err
	}
	return ctr, nil
}

// example usage: "dagger call run --helix-credentials ~/.helix/credentials --prompt hi"
func (m *Helix) Run(ctx context.Context, prompt string, helixCredentials *File) (string, error) {
	ctr, err := m.HelixCli(ctx, helixCredentials)
	if err != nil {
		return "", err
	}
	// TODO: json output from run cmd?
	return ctr.
		WithEntrypoint([]string{"/helix"}).
		WithExec([]string{"run", "--prompt", prompt}).
		Stdout(ctx)
}
