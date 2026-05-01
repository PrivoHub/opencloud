package middleware

import (
	"context"
	"testing"
	"time"

	gateway "github.com/cs3org/go-cs3apis/cs3/gateway/v1beta1"
	cs3user "github.com/cs3org/go-cs3apis/cs3/identity/user/v1beta1"
	rpcpb "github.com/cs3org/go-cs3apis/cs3/rpc/v1beta1"
	storageprovider "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
	"github.com/jellydator/ttlcache/v3"
	"github.com/opencloud-eu/opencloud/pkg/log"
	"github.com/opencloud-eu/opencloud/services/proxy/pkg/config"
	cs3mocks "github.com/opencloud-eu/reva/v2/tests/cs3mocks/mocks"
	"github.com/opencloud-eu/reva/v2/pkg/rgrpc/todo/pool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

const (
	testQuotaUserID    = "user-quota-1"
	testQuotaSpaceID   = "space-personal-1"
	testQuotaSpaceRoot = "space-root-1"
	testQuotaBytes     = uint64(15 << 30) // 15 GiB
)

func newQuotaResolver(t *testing.T, gwc gateway.GatewayAPIClient, quota uint64, cacheTTL time.Duration) accountResolver {
	t.Helper()
	selectorName := "GatewaySelector" + t.Name()
	gatewaySelector := pool.GetSelector[gateway.GatewayAPIClient](
		"GatewaySelector",
		selectorName,
		func(cc grpc.ClientConnInterface) gateway.GatewayAPIClient { return gwc },
	)
	t.Cleanup(func() { pool.RemoveSelector("GatewaySelector" + selectorName) })

	cache := ttlcache.New(
		ttlcache.WithTTL[string, struct{}](cacheTTL),
		ttlcache.WithDisableTouchOnHit[string, struct{}](),
	)
	go cache.Start()
	t.Cleanup(func() { cache.Stop() })

	return accountResolver{
		logger: log.NewLogger(),
		serviceAccount: config.ServiceAccount{
			ServiceAccountID:     testSvcAccountID,
			ServiceAccountSecret: testSvcAccountSecret,
		},
		gatewaySelector:            gatewaySelector,
		privohubPersonalQuotaBytes: quota,
		privohubQuotaApplyCache:    cache,
	}
}

func mockUser() *cs3user.User {
	return &cs3user.User{
		Id: &cs3user.UserId{
			Idp:      testIdP,
			OpaqueId: testQuotaUserID,
		},
		Username: "quota-user",
	}
}

func mockListResponse(found bool) *storageprovider.ListStorageSpacesResponse {
	resp := &storageprovider.ListStorageSpacesResponse{
		Status: &rpcpb.Status{Code: rpcpb.Code_CODE_OK},
	}
	if found {
		resp.StorageSpaces = []*storageprovider.StorageSpace{{
			Id:        &storageprovider.StorageSpaceId{OpaqueId: testQuotaSpaceID},
			Root:      &storageprovider.ResourceId{StorageId: testQuotaSpaceRoot, OpaqueId: testQuotaSpaceRoot},
			SpaceType: "personal",
		}}
	}
	return resp
}

func expectAuthenticate(gwc *cs3mocks.GatewayAPIClient) {
	gwc.On("Authenticate", mock.Anything, &gateway.AuthenticateRequest{
		Type:         "serviceaccounts",
		ClientId:     testSvcAccountID,
		ClientSecret: testSvcAccountSecret,
	}).Return(&gateway.AuthenticateResponse{
		Status: &rpcpb.Status{Code: rpcpb.Code_CODE_OK},
		Token:  testSvcAccountToken,
	}, nil)
}

func TestApplyPersonalQuota_WritesQuotaOnce(t *testing.T) {
	gwc := &cs3mocks.GatewayAPIClient{}
	expectAuthenticate(gwc)
	gwc.On("ListStorageSpaces", mock.Anything, mock.Anything).Return(mockListResponse(true), nil)

	var capturedQuota uint64
	gwc.On("UpdateStorageSpace", mock.Anything, mock.MatchedBy(func(req *storageprovider.UpdateStorageSpaceRequest) bool {
		capturedQuota = req.GetStorageSpace().GetQuota().GetQuotaMaxBytes()
		return req.GetStorageSpace().GetId().GetOpaqueId() == testQuotaSpaceID
	})).Return(&storageprovider.UpdateStorageSpaceResponse{
		Status:       &rpcpb.Status{Code: rpcpb.Code_CODE_OK},
		StorageSpace: &storageprovider.StorageSpace{Id: &storageprovider.StorageSpaceId{OpaqueId: testQuotaSpaceID}},
	}, nil)

	m := newQuotaResolver(t, gwc, testQuotaBytes, 5*time.Minute)
	if err := m.applyPersonalQuota(context.Background(), mockUser()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assert.Equal(t, testQuotaBytes, capturedQuota, "QuotaMaxBytes should match configured value")
	gwc.AssertNumberOfCalls(t, "UpdateStorageSpace", 1)
}

func TestApplyPersonalQuota_CachePreventsRepeatedWrites(t *testing.T) {
	gwc := &cs3mocks.GatewayAPIClient{}
	expectAuthenticate(gwc)
	gwc.On("ListStorageSpaces", mock.Anything, mock.Anything).Return(mockListResponse(true), nil)
	gwc.On("UpdateStorageSpace", mock.Anything, mock.Anything).Return(&storageprovider.UpdateStorageSpaceResponse{
		Status:       &rpcpb.Status{Code: rpcpb.Code_CODE_OK},
		StorageSpace: &storageprovider.StorageSpace{Id: &storageprovider.StorageSpaceId{OpaqueId: testQuotaSpaceID}},
	}, nil)

	m := newQuotaResolver(t, gwc, testQuotaBytes, 5*time.Minute)
	user := mockUser()
	for i := 0; i < 5; i++ {
		if err := m.applyPersonalQuota(context.Background(), user); err != nil {
			t.Fatalf("call %d unexpected error: %v", i, err)
		}
	}
	gwc.AssertNumberOfCalls(t, "UpdateStorageSpace", 1)
}

func TestApplyPersonalQuota_NoSpaceIsSilentNoop(t *testing.T) {
	gwc := &cs3mocks.GatewayAPIClient{}
	expectAuthenticate(gwc)
	gwc.On("ListStorageSpaces", mock.Anything, mock.Anything).Return(mockListResponse(false), nil)

	m := newQuotaResolver(t, gwc, testQuotaBytes, 5*time.Minute)
	if err := m.applyPersonalQuota(context.Background(), mockUser()); err != nil {
		t.Fatalf("unexpected error when personal space is missing: %v", err)
	}
	gwc.AssertNotCalled(t, "UpdateStorageSpace")

	// cache must NOT be populated for the missing-space case so a subsequent
	// request retries the lookup.
	if m.privohubQuotaApplyCache.Has(testQuotaUserID) {
		t.Fatal("cache should not be populated when personal space was not found")
	}
}
