package extauthz

import (
	"context"
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/structpb"
)

func init() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	Init(logger, &Config{Port: 10001})
}

// TestCheck_NoAttributes_Denied verifies that a request with no attributes is denied.
func TestCheck_NoAttributes_Denied(t *testing.T) {
	s := &server{}
	resp, err := s.Check(context.Background(), &authv3.CheckRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status.Code != int32(codes.Unauthenticated) {
		t.Errorf("expected Unauthenticated(%d), got %d", int32(codes.Unauthenticated), resp.Status.Code)
	}
	denied := resp.GetDeniedResponse()
	if denied == nil {
		t.Fatal("expected DeniedResponse, got nil")
	}
	if !hasXFailHeader(denied) {
		t.Error("expected x-fail: auth header in denied response")
	}
	if denied.Body != "{}" {
		t.Errorf("expected body '{}', got %q", denied.Body)
	}
}

// TestCheck_JwtNsAbsent_Denied verifies denial when jwt_authn namespace is missing.
func TestCheck_JwtNsAbsent_Denied(t *testing.T) {
	s := &server{}
	req := &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			MetadataContext: &corev3.Metadata{
				FilterMetadata: map[string]*structpb.Struct{},
			},
		},
	}
	resp, err := s.Check(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status.Code != int32(codes.Unauthenticated) {
		t.Errorf("expected Unauthenticated, got %d", resp.Status.Code)
	}
	if !hasXFailHeader(resp.GetDeniedResponse()) {
		t.Error("expected x-fail: auth header")
	}
}

// TestCheck_JwtNsPresent_NoProviderKey_Denied verifies denial when namespace present but no provider payload.
func TestCheck_JwtNsPresent_NoProviderKey_Denied(t *testing.T) {
	s := &server{}
	req := &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			MetadataContext: &corev3.Metadata{
				FilterMetadata: map[string]*structpb.Struct{
					"envoy.filters.http.jwt_authn": {},
				},
			},
		},
	}
	resp, err := s.Check(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status.Code != int32(codes.Unauthenticated) {
		t.Errorf("expected Unauthenticated, got %d", resp.Status.Code)
	}
}

// TestCheck_Provider1_Allowed verifies that a request with provider_okta_1 payload is allowed.
func TestCheck_Provider1_Allowed(t *testing.T) {
	s := &server{}
	payload, _ := structpb.NewStruct(map[string]interface{}{
		"sub": "user1",
		"iss": "https://integrator-1045327.okta.com/oauth2/default",
	})
	req := &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			MetadataContext: &corev3.Metadata{
				FilterMetadata: map[string]*structpb.Struct{
					"envoy.filters.http.jwt_authn": {
						Fields: map[string]*structpb.Value{
							"provider_okta_1": structpb.NewStructValue(payload),
						},
					},
				},
			},
		},
	}
	resp, err := s.Check(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status.Code != int32(codes.OK) {
		t.Errorf("expected OK, got %d", resp.Status.Code)
	}
	if resp.GetOkResponse() == nil {
		t.Error("expected OkResponse, got nil")
	}
}

// TestCheck_Provider2_Allowed verifies that a request with provider_okta_2 payload is allowed.
func TestCheck_Provider2_Allowed(t *testing.T) {
	s := &server{}
	payload, _ := structpb.NewStruct(map[string]interface{}{
		"sub": "user2",
	})
	req := &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			MetadataContext: &corev3.Metadata{
				FilterMetadata: map[string]*structpb.Struct{
					"envoy.filters.http.jwt_authn": {
						Fields: map[string]*structpb.Value{
							"provider_okta_2": structpb.NewStructValue(payload),
						},
					},
				},
			},
		},
	}
	resp, err := s.Check(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status.Code != int32(codes.OK) {
		t.Errorf("expected OK, got %d", resp.Status.Code)
	}
}

func hasXFailHeader(denied *authv3.DeniedHttpResponse) bool {
	if denied == nil {
		return false
	}
	for _, h := range denied.Headers {
		if h.Header.Key == "x-fail" && h.Header.Value == "auth" {
			return true
		}
	}
	return false
}
