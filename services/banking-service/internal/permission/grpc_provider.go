package permission

import (
	"common/pkg/auth"
	"common/pkg/jwt"
	"common/pkg/pb"
	perm "common/pkg/permission"
	"context"
)

type GrpcPermissionProvider struct {
	client pb.PermissionServiceClient
}

func NewGrpcPermissionProvider(client pb.PermissionServiceClient) *GrpcPermissionProvider {
	return &GrpcPermissionProvider{client: client}
}

func (p *GrpcPermissionProvider) GetPermissions(ctx context.Context, claims *jwt.Claims) ([]perm.Permission, error) {
	// TODO: Decide on this when we scope whether clients will also have permissions
	if auth.IdentityType(claims.IdentityType) != auth.IdentityEmployee {
		return []perm.Permission{}, nil
	}

	req := &pb.GetPermissionsRequest{
		IdentityId:   uint64(claims.IdentityID),
		IdentityType: claims.IdentityType,
		SubjectId:    uint64(*claims.EmployeeID),
	}

	resp, err := p.client.GetPermissions(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]perm.Permission, 0, len(resp.GetPermissions()))
	for _, permName := range resp.GetPermissions() {
		result = append(result, perm.Permission(permName))
	}

	return result, nil
}
