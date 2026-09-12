package auth

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

type TokenVerifier interface {
	Verify(
		ctx context.Context,
		rawToken string,
	) (Identity, error)
}

type TokenVerifierConfig struct {
	Issuer   string
	ClientID string
	TokenUse string
}

type cognitoClaims struct {
	jwt.RegisteredClaims

	Email    string `json:"email"`
	TokenUse string `json:"token_use"`
	ClientID string `json:"client_id"`
}

type unverifiedTokenVerifier struct{}

type cognitoTokenVerifier struct {
	issuer   string
	clientID string
	tokenUse string
	keys     keyfunc.Keyfunc
}

func NewTokenVerifier(
	ctx context.Context,
	cfg TokenVerifierConfig,
) (TokenVerifier, error) {
	issuer := strings.TrimRight(
		strings.TrimSpace(cfg.Issuer),
		"/",
	)

	//
	// Local development mode.
	//
	// If no issuer is configured, JWT signatures are not
	// verified. We only decode the token and extract email.
	//
	if issuer == "" {
		return &unverifiedTokenVerifier{}, nil
	}

	clientID := strings.TrimSpace(
		cfg.ClientID,
	)

	if clientID == "" {
		return nil, fmt.Errorf(
			"JWT_CLIENT_ID is required when JWT_ISSUER is configured",
		)
	}

	tokenUse := strings.ToLower(
		strings.TrimSpace(cfg.TokenUse),
	)

	if tokenUse == "" {
		tokenUse = "id"
	}

	if tokenUse != "id" &&
		tokenUse != "access" {
		return nil, fmt.Errorf(
			"JWT_TOKEN_USE must be either id or access",
		)
	}

	issuerURL, err := url.Parse(issuer)
	if err != nil {
		return nil, fmt.Errorf(
			"parse JWT issuer: %w",
			err,
		)
	}

	if issuerURL.Scheme != "https" ||
		issuerURL.Host == "" {
		return nil, fmt.Errorf(
			"JWT_ISSUER must be a valid HTTPS URL",
		)
	}

	jwksURL :=
		issuer + "/.well-known/jwks.json"

	keys, err := keyfunc.NewDefaultCtx(
		ctx,
		[]string{
			jwksURL,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"initialize JWT JWKS: %w",
			err,
		)
	}

	return &cognitoTokenVerifier{
		issuer:   issuer,
		clientID: clientID,
		tokenUse: tokenUse,
		keys:     keys,
	}, nil
}

func (v *unverifiedTokenVerifier) Verify(
	_ context.Context,
	rawToken string,
) (Identity, error) {
	claims := &cognitoClaims{}

	parser := jwt.NewParser()

	_, _, err := parser.ParseUnverified(
		rawToken,
		claims,
	)
	if err != nil {
		return Identity{}, fmt.Errorf(
			"decode JWT: %w",
			err,
		)
	}

	email := strings.TrimSpace(
		claims.Email,
	)

	if email == "" {
		return Identity{}, fmt.Errorf(
			"JWT does not contain email claim",
		)
	}

	return Identity{
		Email: email,
	}, nil
}

func (v *cognitoTokenVerifier) Verify(
	ctx context.Context,
	rawToken string,
) (Identity, error) {
	claims := &cognitoClaims{}

	options := []jwt.ParserOption{
		jwt.WithValidMethods(
			[]string{
				jwt.SigningMethodRS256.Alg(),
			},
		),

		jwt.WithIssuer(
			v.issuer,
		),

		jwt.WithExpirationRequired(),
	}

	//
	// Cognito ID tokens use "aud" for the app client ID.
	//
	if v.tokenUse == "id" {
		options = append(
			options,
			jwt.WithAudience(
				v.clientID,
			),
		)
	}

	token, err := jwt.ParseWithClaims(
		rawToken,
		claims,
		v.keys.KeyfuncCtx(ctx),
		options...,
	)
	if err != nil {
		return Identity{}, fmt.Errorf(
			"verify JWT: %w",
			err,
		)
	}

	if !token.Valid {
		return Identity{}, fmt.Errorf(
			"JWT is not valid",
		)
	}

	if claims.TokenUse != v.tokenUse {
		return Identity{}, fmt.Errorf(
			"unexpected JWT token_use %q",
			claims.TokenUse,
		)
	}

	//
	// Cognito access tokens use "client_id" instead
	// of "aud" for the app client ID.
	//
	if v.tokenUse == "access" &&
		claims.ClientID != v.clientID {
		return Identity{}, fmt.Errorf(
			"unexpected JWT client_id",
		)
	}

	email := strings.TrimSpace(
		claims.Email,
	)

	if email == "" {
		return Identity{}, fmt.Errorf(
			"JWT does not contain email claim",
		)
	}

	return Identity{
		Email: email,
	}, nil
}
