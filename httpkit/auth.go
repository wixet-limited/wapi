package httpkit

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	appauth "github.com/wixet-limited/wapi/auth"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	oapimiddleware "github.com/oapi-codegen/nethttp-middleware"
)

func OpenAPIValidation(
	spec *openapi3.T,
	verifier appauth.TokenVerifier,
) func(http.Handler) http.Handler {
	spec.Servers = nil

	return oapimiddleware.OapiRequestValidatorWithOptions(
		spec,
		&oapimiddleware.Options{
			Options: openapi3filter.Options{
				AuthenticationFunc: func(
					ctx context.Context,
					input *openapi3filter.AuthenticationInput,
				) error {
					return authenticateBearer(
						ctx,
						input,
						verifier,
					)
				},
			},

			Skipper: func(
				r *http.Request,
			) bool {
				return r.URL.Path == "/livez" ||
					r.URL.Path == "/readyz" ||
					r.Method == http.MethodOptions
			},

			ErrorHandlerWithOpts: openAPIErrorHandler,
		},
	)
}

func authenticateBearer(
	ctx context.Context,
	input *openapi3filter.AuthenticationInput,
	verifier appauth.TokenVerifier,
) error {
	if input.SecuritySchemeName !=
		"bearerAuth" {
		return fmt.Errorf(
			"unsupported security scheme %q",
			input.SecuritySchemeName,
		)
	}

	request :=
		input.RequestValidationInput.Request

	header := strings.TrimSpace(
		request.Header.Get(
			"Authorization",
		),
	)

	scheme, token, found :=
		strings.Cut(
			header,
			" ",
		)

	if !found ||
		!strings.EqualFold(
			scheme,
			"Bearer",
		) ||
		strings.TrimSpace(token) == "" {
		return errors.New(
			"missing bearer token",
		)
	}

	identity, err := verifier.Verify(
		ctx,
		strings.TrimSpace(token),
	)
	if err != nil {
		return fmt.Errorf(
			"invalid bearer token: %w",
			err,
		)
	}

	requestContext :=
		appauth.WithIdentity(
			request.Context(),
			identity,
		)

	//
	// AuthenticationFunc cannot return a new context,
	// therefore update the request used downstream.
	//
	*request =
		*request.WithContext(
			requestContext,
		)

	return nil
}

func openAPIErrorHandler(
	_ context.Context,
	err error,
	w http.ResponseWriter,
	_ *http.Request,
	options oapimiddleware.ErrorHandlerOpts,
) {
	var securityError *openapi3filter.SecurityRequirementsError

	if errors.As(
		err,
		&securityError,
	) {
		WriteProblem(
			w,
			http.StatusUnauthorized,
			"Unauthorized",
			"A valid bearer token is required.",
		)

		return
	}

	if options.MatchedRoute == nil {
		WriteProblem(
			w,
			http.StatusNotFound,
			"Not found",
			"The requested resource does not exist.",
		)

		return
	}

	WriteProblem(
		w,
		options.StatusCode,
		"Invalid request",
		"The request does not match the API contract.",
	)
}
