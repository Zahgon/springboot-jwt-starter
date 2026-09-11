package security_test

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/bfwg/springboot-jwt-starter/internal/clock"
	"github.com/bfwg/springboot-jwt-starter/internal/model"
	"github.com/bfwg/springboot-jwt-starter/internal/security"
)

const testUsername = "testUser"

// dateUtil helpers mirror the assertj DateUtil the original's fixture used.
func now() time.Time       { return time.Now() }
func yesterday() time.Time { return now().AddDate(0, 0, -1) }
func tomorrow() time.Time  { return now().AddDate(0, 0, 1) }

// newTokenHelper builds the fixture the original set up in @Before: a 10 second
// lifetime, a pinned secret, and a scriptable time source.
func newTokenHelper(t *testing.T) (*security.TokenHelper, *clock.Mock) {
	t.Helper()
	timeProviderMock := clock.NewMock()
	helper := security.NewTokenHelper(
		"springboot-jwt-demo",
		"Y0[yCzX7Ym;${,[+hj*(E*erX:-%-JU=K!}Sp/m:$gZ.D[EW,fQec3Ha1.rwy)TX",
		10,
		"Authorization",
		timeProviderMock,
	)
	return helper, timeProviderMock
}

func TestGenerateTokenGeneratesDifferentTokensForDifferentCreationDates(t *testing.T) {
	tokenHelper, timeProviderMock := newTokenHelper(t)
	timeProviderMock.Returns(yesterday(), now())

	token := tokenHelper.GenerateToken(testUsername)
	laterToken := tokenHelper.GenerateToken(testUsername)

	if token == laterToken {
		t.Fatalf("expected different tokens for different creation dates, got %q twice", token)
	}
}

func TestTokenShouldExpire(t *testing.T) {
	tokenHelper, timeProviderMock := newTokenHelper(t)
	beforeSomeTime := now().Add(-20 * time.Second)
	timeProviderMock.Returns(beforeSomeTime)

	userDetails := &mockUser{username: testUsername}

	mobileToken := tokenHelper.GenerateToken(testUsername)
	if tokenHelper.ValidateToken(mobileToken, userDetails) {
		t.Fatal("expected an expired token to fail validation")
	}
}

func TestGetUsernameFromToken(t *testing.T) {
	tokenHelper, timeProviderMock := newTokenHelper(t)
	timeProviderMock.Returns(now())

	token := tokenHelper.GenerateToken(testUsername)

	if got := tokenHelper.GetUsernameFromToken(token); got != testUsername {
		t.Fatalf("username from token = %q, want %q", got, testUsername)
	}
}

func TestGetCreatedDateFromToken(t *testing.T) {
	tokenHelper, timeProviderMock := newTokenHelper(t)
	instant := now()
	timeProviderMock.Returns(instant)

	token := tokenHelper.GenerateToken(testUsername)

	got := tokenHelper.GetIssuedAtDateFromToken(token)
	if delta := got.Sub(instant); delta > time.Minute || delta < -time.Minute {
		t.Fatalf("issued-at %v is not in the same minute window as %v", got, instant)
	}
}

func TestExpiredTokenCannotBeRefreshed(t *testing.T) {
	tokenHelper, timeProviderMock := newTokenHelper(t)
	timeProviderMock.Returns(yesterday())

	token := tokenHelper.GenerateToken(testUsername)
	// Must not panic; the original asserts only that refreshing is survivable.
	tokenHelper.RefreshToken(token)
}

func TestChangedPasswordCannotBeRefreshed(t *testing.T) {
	tokenHelper, timeProviderMock := newTokenHelper(t)
	timeProviderMock.Returns(now())

	user := &mockUser{lastPasswordResetDate: model.Timestamp(tomorrow())}
	token := tokenHelper.GenerateToken(testUsername)

	if tokenHelper.ValidateToken(token, user) {
		t.Fatal("expected validation to fail after a password change")
	}
}

func TestCanRefreshToken(t *testing.T) {
	tokenHelper, timeProviderMock := newTokenHelper(t)
	timeProviderMock.Returns(now(), tomorrow())

	firstToken := tokenHelper.GenerateToken(testUsername)
	refreshedToken := tokenHelper.RefreshToken(firstToken)

	firstTokenDate := tokenHelper.GetIssuedAtDateFromToken(firstToken)
	refreshedTokenDate := tokenHelper.GetIssuedAtDateFromToken(refreshedToken)

	if !firstTokenDate.Before(refreshedTokenDate) {
		t.Fatalf("first issued-at %v is not before refreshed issued-at %v",
			firstTokenDate, refreshedTokenDate)
	}
}

// --- Tests for behaviour the original delegated to jjwt, and this port now
// --- implements itself.

func decodeSegment(t *testing.T, segment string) map[string]any {
	t.Helper()
	raw, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		t.Fatalf("decoding segment: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshalling segment: %v", err)
	}
	return out
}

func TestTokenHeaderIsBareHS512(t *testing.T) {
	tokenHelper, timeProviderMock := newTokenHelper(t)
	timeProviderMock.Returns(now())

	parts := strings.Split(tokenHelper.GenerateToken(testUsername), ".")
	if len(parts) != 3 {
		t.Fatalf("expected a three-part compact JWS, got %d parts", len(parts))
	}

	header := decodeSegment(t, parts[0])
	if header["alg"] != "HS512" {
		t.Errorf(`header alg = %v, want "HS512"`, header["alg"])
	}
	if _, present := header["typ"]; present {
		t.Error("header must not carry a typ member")
	}
	if len(header) != 1 {
		t.Errorf("header has %d members, want exactly 1", len(header))
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatalf("decoding signature: %v", err)
	}
	if len(signature) != 64 {
		t.Errorf("signature is %d bytes, want 64 for HS512", len(signature))
	}
}

func TestTokenClaimSetIsClosed(t *testing.T) {
	tokenHelper, timeProviderMock := newTokenHelper(t)
	instant := now()
	timeProviderMock.Returns(instant)

	parts := strings.Split(tokenHelper.GenerateToken(testUsername), ".")
	claims := decodeSegment(t, parts[1])

	for _, key := range []string{"iss", "sub", "iat", "exp"} {
		if _, present := claims[key]; !present {
			t.Errorf("claim %q is missing", key)
		}
	}
	if len(claims) != 4 {
		t.Errorf("claim set has %d members %v, want exactly 4", len(claims), claims)
	}
	if claims["iss"] != "springboot-jwt-demo" {
		t.Errorf(`iss = %v, want "springboot-jwt-demo"`, claims["iss"])
	}
	if claims["sub"] != testUsername {
		t.Errorf("sub = %v, want %q", claims["sub"], testUsername)
	}
	iat, exp := claims["iat"].(float64), claims["exp"].(float64)
	if exp-iat != 10 {
		t.Errorf("exp - iat = %v, want the configured 10 seconds", exp-iat)
	}
	if int64(iat) != instant.Unix() {
		t.Errorf("iat = %v, want %v seconds since epoch", int64(iat), instant.Unix())
	}
}

func TestValidateTokenRejectsTokenIssuedBeforeLastPasswordReset(t *testing.T) {
	tokenHelper, timeProviderMock := newTokenHelper(t)
	instant := now()
	timeProviderMock.Returns(instant)

	token := tokenHelper.GenerateToken(testUsername)

	// Same username, so only the password-reset instant decides the outcome.
	beforeReset := &mockUser{
		username:              testUsername,
		lastPasswordResetDate: model.Timestamp(instant.Add(time.Hour)),
	}
	if tokenHelper.ValidateToken(token, beforeReset) {
		t.Error("token issued before the last password reset must not validate")
	}

	afterReset := &mockUser{
		username:              testUsername,
		lastPasswordResetDate: model.Timestamp(instant.Add(-time.Hour)),
	}
	if !tokenHelper.ValidateToken(token, afterReset) {
		t.Error("token issued after the last password reset must validate")
	}
}

func TestValidateTokenRejectsMismatchedUsername(t *testing.T) {
	tokenHelper, timeProviderMock := newTokenHelper(t)
	timeProviderMock.Returns(now())

	token := tokenHelper.GenerateToken(testUsername)

	if tokenHelper.ValidateToken(token, newUserDetailsDummy("someoneElse")) {
		t.Error("a token naming a different subject must not validate")
	}
	if !tokenHelper.ValidateToken(token, newUserDetailsDummy(testUsername)) {
		t.Error("a token naming this subject must validate")
	}
}

func TestTokenSignedWithADifferentSecretIsRejected(t *testing.T) {
	tokenHelper, timeProviderMock := newTokenHelper(t)
	timeProviderMock.Returns(now())
	token := tokenHelper.GenerateToken(testUsername)

	other := security.NewTokenHelper("springboot-jwt-demo", "a different secret entirely",
		10, "Authorization", clock.NewMock(now()))

	if got := other.GetUsernameFromToken(token); got != "" {
		t.Errorf("username from a foreign-signed token = %q, want empty", got)
	}
	if other.RefreshToken(token) != "" {
		t.Error("a foreign-signed token must not be refreshable")
	}
}

func TestRefreshTokenPreservesIssuerAndSubject(t *testing.T) {
	tokenHelper, timeProviderMock := newTokenHelper(t)
	timeProviderMock.Returns(now(), now().Add(time.Second))

	refreshed := tokenHelper.RefreshToken(tokenHelper.GenerateToken(testUsername))
	if refreshed == "" {
		t.Fatal("expected a refreshed token")
	}

	claims := decodeSegment(t, strings.Split(refreshed, ".")[1])
	if claims["iss"] != "springboot-jwt-demo" {
		t.Errorf("refreshed iss = %v, want the original issuer", claims["iss"])
	}
	if claims["sub"] != testUsername {
		t.Errorf("refreshed sub = %v, want %q", claims["sub"], testUsername)
	}
}

// Claims implements jwt.Claims. The library reads these accessors when the
// matching validation option is enabled, so a wrong one would fail silently
// today and break the moment issuer or audience checking is turned on.
func TestClaimsImplementsJWTClaims(t *testing.T) {
	claims := security.Claims{
		Issuer:   "springboot-jwt-demo",
		Subject:  testUsername,
		IssuedAt: 1_700_000_000,
		Expiry:   1_700_000_300,
	}

	issuer, err := claims.GetIssuer()
	if err != nil || issuer != "springboot-jwt-demo" {
		t.Errorf("GetIssuer = %q, %v; want the configured issuer", issuer, err)
	}

	subject, err := claims.GetSubject()
	if err != nil || subject != testUsername {
		t.Errorf("GetSubject = %q, %v; want %q", subject, err, testUsername)
	}

	issuedAt, err := claims.GetIssuedAt()
	if err != nil || issuedAt == nil || issuedAt.Unix() != 1_700_000_000 {
		t.Errorf("GetIssuedAt = %v, %v; want 1700000000", issuedAt, err)
	}

	expiry, err := claims.GetExpirationTime()
	if err != nil || expiry == nil || expiry.Unix() != 1_700_000_300 {
		t.Errorf("GetExpirationTime = %v, %v; want 1700000300", expiry, err)
	}

	// The original sets neither nbf nor aud.
	notBefore, err := claims.GetNotBefore()
	if err != nil || notBefore != nil {
		t.Errorf("GetNotBefore = %v, %v; want nil", notBefore, err)
	}
	audience, err := claims.GetAudience()
	if err != nil || audience != nil {
		t.Errorf("GetAudience = %v, %v; want nil", audience, err)
	}
}
