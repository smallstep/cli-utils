package token

import (
	"encoding/base64"
	"reflect"
	"testing"
	"time"

	"github.com/smallstep/assert"
	"go.step.sm/crypto/jose"
	"go.step.sm/crypto/pemutil"
)

func TestOptions(t *testing.T) {
	var (
		empty = new(Claims)
		now   = time.Now()
	)

	temp := timeNowUTC
	timeNowUTC = func() time.Time {
		return now.UTC()
	}
	t.Cleanup(func() {
		timeNowUTC = temp
	})

	certs, err := pemutil.ReadCertificateBundle("./testdata/foo.crt")
	assert.FatalError(t, err)
	certStrs := make([]string, len(certs))
	for i, cert := range certs {
		certStrs[i] = base64.StdEncoding.EncodeToString(cert.Raw)
	}

	x5cKey, err := pemutil.Read("./testdata/foo.key")
	assert.FatalError(t, err)

	tests := []struct {
		name    string
		option  Options
		want    *Claims
		wantErr bool
	}{
		{"WithClaim ok", WithClaim("name", "foo"), &Claims{ExtraClaims: map[string]interface{}{"name": "foo"}}, false},
		{"WithClaim fail", WithClaim("", "foo"), empty, true},
		{"WithRootCA ok", WithRootCA("testdata/ca.crt"), &Claims{ExtraClaims: map[string]interface{}{"sha": "6908751f68290d4573ae0be39a98c8b9b7b7d4e8b2a6694b7509946626adfe98"}}, false},
		{"WithRootCA fail", WithRootCA("not-exists"), empty, true},
		{"WithIssuedAt", WithIssuedAt(now), &Claims{Claims: jose.Claims{IssuedAt: jose.NewNumericDate(now)}}, false},
		{"WithIssuedAt zero", WithIssuedAt(time.Time{}), &Claims{Claims: jose.Claims{IssuedAt: nil}}, false},
		{"WithValidity ok", WithValidity(now, now.Add(5*time.Minute)), &Claims{Claims: jose.Claims{NotBefore: jose.NewNumericDate(now), Expiry: jose.NewNumericDate(now.Add(5 * time.Minute))}}, false},
		{"WithValidity zero nbf", WithValidity(time.Time{}, now), &Claims{Claims: jose.Claims{NotBefore: nil, Expiry: jose.NewNumericDate(now)}}, false},
		{"WithValidity zero exp", WithValidity(now, time.Time{}), &Claims{Claims: jose.Claims{NotBefore: jose.NewNumericDate(now), Expiry: nil}}, false},
		{"WithValidity zero", WithValidity(time.Time{}, time.Time{}), &Claims{Claims: jose.Claims{NotBefore: nil, Expiry: nil}}, false},
		{"WithValidity fail expired", WithValidity(now, now.Add(-time.Second)), &Claims{Claims: jose.Claims{}}, true},
		{"WithValidity fail delay validity", WithValidity(now.Add(MaxValidityDelay+1), now.Add(MaxValidityDelay+5*time.Minute)), &Claims{Claims: jose.Claims{}}, true},
		{"WithValidity fail min validity", WithValidity(now, now.Add(MinValidity-1)), &Claims{Claims: jose.Claims{}}, true},
		{"WithValidity fail max validity", WithValidity(now, now.Add(MaxValidity+1)), &Claims{Claims: jose.Claims{}}, true},
		{"WithRootCA expired", WithValidity(now, now.Add(-1*time.Second)), empty, true},
		{"WithRootCA long delay", WithValidity(now.Add(MaxValidityDelay+time.Minute), now.Add(MaxValidityDelay+10*time.Minute)), empty, true},
		{"WithRootCA min validity ok", WithValidity(now, now.Add(MinValidity)), &Claims{Claims: jose.Claims{NotBefore: jose.NewNumericDate(now), Expiry: jose.NewNumericDate(now.Add(MinValidity))}}, false},
		{"WithRootCA min validity fail", WithValidity(now, now.Add(MinValidity-time.Second)), empty, true},
		{"WithRootCA max validity ok", WithValidity(now, now.Add(MaxValidity)), &Claims{Claims: jose.Claims{NotBefore: jose.NewNumericDate(now), Expiry: jose.NewNumericDate(now.Add(MaxValidity))}}, false},
		{"WithRootCA max validity fail", WithValidity(now, now.Add(MaxValidity+time.Second)), empty, true},
		{"WithIssuer ok", WithIssuer("value"), &Claims{Claims: jose.Claims{Issuer: "value"}}, false},
		{"WithIssuer fail", WithIssuer(""), empty, true},
		{"WithSubject ok", WithSubject("value"), &Claims{Claims: jose.Claims{Subject: "value"}}, false},
		{"WithSubject fail", WithSubject(""), empty, true},
		{"WithAudience ok", WithAudience("value"), &Claims{Claims: jose.Claims{Audience: jose.Audience{"value"}}}, false},
		{"WithAudience fail", WithAudience(""), empty, true},
		{"WithJWTID ok", WithJWTID("value"), &Claims{Claims: jose.Claims{ID: "value"}}, false},
		{"WithJWTID fail", WithJWTID(""), empty, true},
		{"WithKid ok", WithKid("value"), &Claims{ExtraHeaders: map[string]interface{}{"kid": "value"}}, false},
		{"WithKid fail", WithKid(""), empty, true},
		{"WithSHA ok", WithSHA("6908751f68290d4573ae0be39a98c8b9b7b7d4e8b2a6694b7509946626adfe98"), &Claims{ExtraClaims: map[string]interface{}{"sha": "6908751f68290d4573ae0be39a98c8b9b7b7d4e8b2a6694b7509946626adfe98"}}, false},
		{"WithX5CCerts ok", WithX5CCerts(certStrs), &Claims{ExtraHeaders: map[string]interface{}{"x5c": certStrs}}, false},
		{"WithX5CFile ok", WithX5CFile("./testdata/foo.crt", x5cKey), &Claims{ExtraHeaders: map[string]interface{}{"x5c": certStrs}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claim := new(Claims)
			err := tt.option(claim)
			if (err != nil) != tt.wantErr {
				t.Errorf("Options error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(claim, tt.want) {
				t.Errorf("Options claims = %v, want %v", claim, tt.want)
			}
		})
	}
}
