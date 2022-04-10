package gcp

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2/google"
)

func NewAuthenticatedClient(scopes []string) (context.Context, *http.Client, error) {
	ctx := context.Background()

	/*
		// TODO(rjk): Fix this.
		// It's commented out. What does it do?
		if *debug {
			ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
				Transport: gcp.NewTransport(http.DefaultTransport),
			})
		}
	*/
	client, err := google.DefaultClient(ctx, strings.Join(scopes, " "))
	if err != nil {
		return nil, nil, fmt.Errorf("can't make oauth client: %v", err)
	}

	return ctx, client, nil
}
