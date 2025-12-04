package auth

// type Provider interface {
// 	ID() string

// 	// 1. I want to login so where do I go?
// 	LoginURL(ctx context.Context, opts LoginOptions) (string, error)

// 	// 2. I’ve been redirected back with ?code=... – finish login.
// 	ExchangeCode(ctx context.Context, code, redirectURI string) (*AuthResult, error)

// 	// 3. Are these credentials valid?
// 	//    You pass tokens you already have (from cookies / headers).
// 	Validate(ctx context.Context, accessToken, idToken string) (*AuthResult, error)

// 	// 4. I need to use my refresh token.
// 	Refresh(ctx context.Context, refreshToken string) (*AuthResult, error)

// 	// 5. I need to log out (build a logout URL, possibly using ID token hint).
// 	LogoutURL(ctx context.Context, opts LogoutOptions) (string, error)

// 	// 6. Optionally, let the provider revoke tokens server-side.
// 	Revoke(ctx context.Context, accessToken, refreshToken string) error
// }
