// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"fmt"
	"net/url"

	"golang.org/x/oauth2/spotify"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/x/httpx"
	"github.com/ory/x/stringslice"
	"github.com/ory/x/stringsx"

	spotifyapi "github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"

	"github.com/ory/herodot"
)

var _ OAuth2Provider = (*ProviderSpotify)(nil)

type ProviderSpotify struct {
	config *Configuration
	reg    Dependencies
}

func NewProviderSpotify(
	config *Configuration,
	reg Dependencies,
) Provider {
	return &ProviderSpotify{
		config: config,
		reg:    reg,
	}
}

func (g *ProviderSpotify) Config() *Configuration {
	return g.config
}

func (g *ProviderSpotify) oauth2(ctx context.Context) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		Endpoint:     spotify.Endpoint,
		Scopes:       g.config.Scope,
		RedirectURL:  g.config.Redir(g.reg.Config().OIDCRedirectURIBase(ctx)),
	}
}

func (g *ProviderSpotify) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return g.oauth2(ctx), nil
}

func (g *ProviderSpotify) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}

func (g *ProviderSpotify) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	grantedScopes := stringsx.Splitx(fmt.Sprintf("%s", exchange.Extra("scope")), " ")
	for _, check := range g.Config().Scope {
		if !stringslice.Has(grantedScopes, check) {
			return nil, errors.WithStack(ErrScopeMissing)
		}
	}

	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(g.config.Redir(g.reg.Config().OIDCRedirectURIBase(ctx))),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate))

	ctx, client := httpx.SetOAuth2(ctx, g.reg.HTTPClient(ctx), auth, exchange)
	spotifyClient := spotifyapi.New(client.HTTPClient)

	user, err := spotifyClient.CurrentUser(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	var userPicture string
	if len(user.Images) > 0 {
		userPicture = user.Images[0].URL
	}

	claims := &Claims{
		Subject:   user.ID,
		Issuer:    spotify.Endpoint.TokenURL,
		Name:      user.DisplayName,
		Nickname:  user.DisplayName,
		Email:     user.Email,
		Picture:   userPicture,
		Profile:   user.ExternalURLs["spotify"],
		Birthdate: user.Birthdate,
	}

	return claims, nil
}
