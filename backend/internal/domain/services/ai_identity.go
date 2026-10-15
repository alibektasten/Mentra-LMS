package services

import (
	"context"

	"github.com/google/uuid"
)

type aiIdentityKey struct{}
type AIIdentity struct {
	OrgID  uuid.UUID
	UserID uuid.UUID
}

func WithAIIdentity(ctx context.Context, orgID, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, aiIdentityKey{}, AIIdentity{OrgID: orgID, UserID: userID})
}

func AIIdentityFromContext(ctx context.Context) (AIIdentity, bool) {
	identity, ok := ctx.Value(aiIdentityKey{}).(AIIdentity)
	return identity, ok && identity.OrgID != uuid.Nil && identity.UserID != uuid.Nil
}
