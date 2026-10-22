package middleware

import (
	"context"
	"fmt"

	cs3user "github.com/cs3org/go-cs3apis/cs3/identity/user/v1beta1"
	rpcpb "github.com/cs3org/go-cs3apis/cs3/rpc/v1beta1"
	storageprovider "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
	"github.com/jellydator/ttlcache/v3"
	"github.com/opencloud-eu/reva/v2/pkg/utils"
)

const privohubPersonalSpaceType = "personal"

// applyPersonalQuota writes m.privohubPersonalQuotaBytes onto the resolved user's personal
// storage space. It is a no-op when the user already had its quota applied within the cache
// window. Errors do not fail the request: quota application is best-effort and retried on
// the next cache miss.
func (m accountResolver) applyPersonalQuota(ctx context.Context, user *cs3user.User) error {
	userID := user.GetId().GetOpaqueId()
	if userID == "" {
		return nil
	}
	if m.privohubQuotaApplyCache != nil && m.privohubQuotaApplyCache.Has(userID) {
		return nil
	}

	gwc, err := m.gatewaySelector.Next()
	if err != nil {
		return fmt.Errorf("could not get gateway client: %w", err)
	}
	authCtx, err := utils.GetServiceUserContextWithContext(ctx, gwc, m.serviceAccount.ServiceAccountID, m.serviceAccount.ServiceAccountSecret)
	if err != nil {
		return fmt.Errorf("could not authenticate service account: %w", err)
	}

	listResp, err := gwc.ListStorageSpaces(authCtx, &storageprovider.ListStorageSpacesRequest{
		Filters: []*storageprovider.ListStorageSpacesRequest_Filter{
			{
				Type: storageprovider.ListStorageSpacesRequest_Filter_TYPE_SPACE_TYPE,
				Term: &storageprovider.ListStorageSpacesRequest_Filter_SpaceType{
					SpaceType: privohubPersonalSpaceType,
				},
			},
			{
				Type: storageprovider.ListStorageSpacesRequest_Filter_TYPE_USER,
				Term: &storageprovider.ListStorageSpacesRequest_Filter_User{
					User: user.GetId(),
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("ListStorageSpaces transport error: %w", err)
	}
	if code := listResp.GetStatus().GetCode(); code != rpcpb.Code_CODE_OK {
		return fmt.Errorf("ListStorageSpaces returned %s: %s", code, listResp.GetStatus().GetMessage())
	}

	spaces := listResp.GetStorageSpaces()
	if len(spaces) == 0 {
		// Personal space not yet created (first-login race), try again on next request.
		return nil
	}
	space := spaces[0]

	updateResp, err := gwc.UpdateStorageSpace(authCtx, &storageprovider.UpdateStorageSpaceRequest{
		StorageSpace: &storageprovider.StorageSpace{
			Id:    space.GetId(),
			Root:  space.GetRoot(),
			Quota: &storageprovider.Quota{QuotaMaxBytes: m.privohubPersonalQuotaBytes},
		},
	})
	if err != nil {
		return fmt.Errorf("UpdateStorageSpace transport error: %w", err)
	}
	if code := updateResp.GetStatus().GetCode(); code != rpcpb.Code_CODE_OK {
		return fmt.Errorf("UpdateStorageSpace returned %s: %s", code, updateResp.GetStatus().GetMessage())
	}

	if m.privohubQuotaApplyCache != nil {
		m.privohubQuotaApplyCache.Set(userID, struct{}{}, ttlcache.DefaultTTL)
	}
	return nil
}
