package extauthz

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	rpcstatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthPb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

const jwtAuthnNamespace = "envoy.filters.http.jwt_authn"

type server struct{}
type healthServer struct{}

func (s *healthServer) Check(_ context.Context, in *healthPb.HealthCheckRequest) (*healthPb.HealthCheckResponse, error) {
	log.Debugf("health check: %s", in.String())
	return &healthPb.HealthCheckResponse{Status: healthPb.HealthCheckResponse_SERVING}, nil
}

func (s *healthServer) Watch(_ *healthPb.HealthCheckRequest, _ healthPb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watch is not implemented")
}

// Check implements the ext_authz Authorization service.
// It reads the jwt_authn filter metadata forwarded by Envoy, logs it, and
// denies requests where no provider JWT payload is present.
func (s *server) Check(_ context.Context, req *authv3.CheckRequest) (*authv3.CheckResponse, error) {
	jwtBody, hasValidJWT := extractJWTMetadata(req)

	log.WithField("jwt_authn", jwtBody).Info("jwt_authn namespace contents")

	if hasValidJWT {
		return &authv3.CheckResponse{
			Status: &rpcstatus.Status{Code: int32(codes.OK)},
			HttpResponse: &authv3.CheckResponse_OkResponse{
				OkResponse: &authv3.OkHttpResponse{},
			},
		}, nil
	}

	return &authv3.CheckResponse{
		Status: &rpcstatus.Status{Code: int32(codes.Unauthenticated)},
		HttpResponse: &authv3.CheckResponse_DeniedResponse{
			DeniedResponse: &authv3.DeniedHttpResponse{
				Status: &typev3.HttpStatus{Code: typev3.StatusCode_Unauthorized},
				Headers: []*corev3.HeaderValueOption{
					{Header: &corev3.HeaderValue{Key: "x-fail", Value: "auth"}},
				},
				Body: jwtBody,
			},
		},
	}, nil
}

// extractJWTMetadata returns the marshaled jwt_authn namespace and whether a valid
// provider payload (provider_okta_1 or provider_okta_2) is present.
func extractJWTMetadata(req *authv3.CheckRequest) (string, bool) {
	if req.Attributes == nil || req.Attributes.MetadataContext == nil {
		return "{}", false
	}

	jwtMeta, ok := req.Attributes.MetadataContext.FilterMetadata[jwtAuthnNamespace]

	if !ok {
		return "not ok", false
	}

	if jwtMeta == nil {
		return "failure case ... this shouldn't happen", false
	}

	body, err := protojson.Marshal(jwtMeta)
	if err != nil {
		log.WithError(err).Warn("failed to marshal jwt_authn metadata")
		body = []byte("{}")
	}

	fields := jwtMeta.GetFields()
	_, hasOkta1 := fields["provider_okta_1"]
	_, hasOkta2 := fields["provider_okta_2"]

	return string(body), hasOkta1 || hasOkta2
}

// Run starts the gRPC server and blocks until SIGINT/SIGTERM.
func Run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	authv3.RegisterAuthorizationServer(grpcServer, &server{})
	healthPb.RegisterHealthServer(grpcServer, &healthServer{})

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.WithError(err).Fatal("gRPC server failed")
		}
	}()

	log.Infof("Listening on port %d", config.Port)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done

	grpcServer.GracefulStop()
	log.Info("Shutdown complete")
	return nil
}
