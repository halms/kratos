// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package link

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/kratos/x"
	"github.com/ory/x/assertx"
	"github.com/ory/x/contextx"
	"github.com/ory/x/sqlcon"
)

func TestPersister(ctx context.Context, p interface {
	persistence.Persister
},
) func(t *testing.T) {
	return func(t *testing.T) {
		nid, p := testhelpers.NewNetworkUnlessExisting(t, ctx, p)

		ctx := contextx.WithConfigValue(ctx, config.ViperKeySecretsDefault, []string{"secret-a", "secret-b"})

		t.Run("token=recovery", func(t *testing.T) {
			newRecoveryToken := func(t *testing.T, email string) (*link.RecoveryToken, *recovery.Flow) {
				var req recovery.Flow
				require.NoError(t, faker.FakeData(&req))
				req.State = flow.StateChooseMethod
				require.NoError(t, p.CreateRecoveryFlow(ctx, &req))

				var i identity.Identity
				require.NoError(t, faker.FakeData(&i))

				address := &identity.RecoveryAddress{Value: email, Via: identity.RecoveryAddressTypeEmail, IdentityID: i.ID}
				i.RecoveryAddresses = append(i.RecoveryAddresses, *address)

				require.NoError(t, p.CreateIdentity(ctx, &i))

				return &link.RecoveryToken{
					Token:           x.NewUUID().String(),
					FlowID:          uuid.NullUUID{UUID: req.ID, Valid: true},
					RecoveryAddress: &i.RecoveryAddresses[0],
					ExpiresAt:       time.Now(),
					IssuedAt:        time.Now(),
					IdentityID:      i.ID,
					TokenType:       link.RecoveryTokenTypeAdmin,
				}, &req
			}

			t.Run("case=should error when the recovery token does not exist", func(t *testing.T) {
				_, err := p.UseRecoveryToken(ctx, x.NewUUID(), "i-do-not-exist")
				require.Error(t, err)
			})

			t.Run("case=should create a new recovery token", func(t *testing.T) {
				token, _ := newRecoveryToken(t, "foo-user@ory.sh")
				require.NoError(t, p.CreateRecoveryToken(ctx, token))
			})

			t.Run("case=should error when token is used with different flow id", func(t *testing.T) {
				token, _ := newRecoveryToken(t, "foo-user1@ory.sh")
				require.NoError(t, p.CreateRecoveryToken(ctx, token))
				_, err := p.UseRecoveryToken(ctx, x.NewUUID(), token.Token)
				require.Error(t, err)
			})

			t.Run("case=should create a recovery token and use it", func(t *testing.T) {
				expected, f := newRecoveryToken(t, "other-user@ory.sh")
				require.NoError(t, p.CreateRecoveryToken(ctx, expected))

				t.Run("not work on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.UseRecoveryToken(ctx, f.ID, expected.Token)
					require.ErrorIs(t, err, sqlcon.ErrNoRows)
				})

				actual, err := p.UseRecoveryToken(ctx, f.ID, expected.Token)
				require.NoError(t, err)
				assert.Equal(t, nid, actual.NID)
				assert.Equal(t, expected.IdentityID, actual.IdentityID)
				assert.NotEqual(t, expected.Token, actual.Token)
				assert.EqualValues(t, expected.FlowID, actual.FlowID)

				t.Run("double spend", func(t *testing.T) {
					_, err = p.UseRecoveryToken(ctx, f.ID, expected.Token)
					require.Error(t, err)
				})
			})

			t.Run("case=update to identity should not invalidate token", func(t *testing.T) {
				expected, f := newRecoveryToken(t, "some-user@ory.sh")

				require.NoError(t, p.CreateRecoveryToken(ctx, expected))
				id, err := p.GetIdentity(ctx, expected.IdentityID, identity.ExpandDefault)
				require.NoError(t, err)
				require.NoError(t, p.UpdateIdentity(ctx, id))

				actual, err := p.UseRecoveryToken(ctx, f.ID, expected.Token)
				require.NoError(t, err)
				assert.Equal(t, nid, actual.NID)
				assert.Equal(t, expected.IdentityID, actual.IdentityID)
				assert.NotEqual(t, expected.Token, actual.Token)
				assert.EqualValues(t, expected.FlowID, actual.FlowID)

				t.Run("double spend", func(t *testing.T) {
					_, err = p.UseRecoveryToken(ctx, f.ID, expected.Token)
					require.Error(t, err)
				})
			})
		})

		t.Run("token=verification", func(t *testing.T) {
			newVerificationToken := func(t *testing.T, email string) (*verification.Flow, *link.VerificationToken) {
				var f verification.Flow
				require.NoError(t, faker.FakeData(&f))
				f.State = flow.StateChooseMethod
				require.NoError(t, p.CreateVerificationFlow(ctx, &f))

				var i identity.Identity
				require.NoError(t, faker.FakeData(&i))

				address := &identity.VerifiableAddress{Value: email, Via: identity.VerifiableAddressTypeEmail}
				i.VerifiableAddresses = append(i.VerifiableAddresses, *address)

				require.NoError(t, p.CreateIdentity(ctx, &i))
				return &f, &link.VerificationToken{
					Token:             x.NewUUID().String(),
					FlowID:            f.ID,
					VerifiableAddress: &i.VerifiableAddresses[0],
					ExpiresAt:         time.Now(),
					IssuedAt:          time.Now(),
				}
			}

			t.Run("case=should error when the verification token does not exist", func(t *testing.T) {
				_, err := p.UseVerificationToken(ctx, x.NewUUID(), "i-do-not-exist")
				require.Error(t, err)
			})

			t.Run("case=should error when the verification token does exist but the flow does not", func(t *testing.T) {
				_, token := newVerificationToken(t, x.NewUUID().String()+"@ory.sh")
				require.NoError(t, p.CreateVerificationToken(ctx, token))
				_, err := p.UseVerificationToken(ctx, x.NewUUID(), token.Token)
				require.Error(t, err)
			})

			t.Run("case=should create a new verification token", func(t *testing.T) {
				_, token := newVerificationToken(t, "foo-user@ory.sh")
				require.NoError(t, p.CreateVerificationToken(ctx, token))
			})

			t.Run("case=should create a verification token and use it", func(t *testing.T) {
				f, expected := newVerificationToken(t, "other-user@ory.sh")
				require.NoError(t, p.CreateVerificationToken(ctx, expected))

				t.Run("not work on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.UseVerificationToken(ctx, f.ID, expected.Token)
					require.ErrorIs(t, err, sqlcon.ErrNoRows)
				})

				actual, err := p.UseVerificationToken(ctx, f.ID, expected.Token)
				require.NoError(t, err)
				assertx.EqualAsJSONExcept(t, expected.VerifiableAddress, actual.VerifiableAddress, []string{"created_at", "updated_at"})
				assert.Equal(t, nid, actual.NID)
				assert.Equal(t, expected.VerifiableAddress.IdentityID, actual.VerifiableAddress.IdentityID)
				assert.NotEqual(t, expected.Token, actual.Token)
				assert.EqualValues(t, expected.FlowID, actual.FlowID)

				_, err = p.UseVerificationToken(ctx, f.ID, expected.Token)
				require.Error(t, err)
			})
		})
	}
}
