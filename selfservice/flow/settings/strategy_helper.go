// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package settings

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

var (
	ErrContinuePreviousAction = errors.New("found prior action")
)

type UpdatePayload interface {
	GetFlowID() uuid.UUID
	SetFlowID(id uuid.UUID)
}

type UpdateContext struct {
	Session  *session.Session
	Flow     *Flow
	toUpdate *identity.Identity
}

func (c *UpdateContext) UpdateIdentity(i *identity.Identity) {
	c.toUpdate = i
}

func (c *UpdateContext) GetIdentityToUpdate() (*identity.Identity, error) {
	if c.toUpdate == nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Could not find a identity to update."))
	}
	return c.toUpdate, nil
}

func (c UpdateContext) GetSessionIdentity() *identity.Identity {
	if c.Session == nil {
		return nil
	}
	return c.Session.Identity
}

func PrepareUpdate(d interface {
	x.LoggingProvider
	continuity.ManagementProvider
}, w http.ResponseWriter, r *http.Request, f *Flow, ss *session.Session, name string, payload UpdatePayload) (*UpdateContext, error) {
	payload.SetFlowID(f.ID)
	c := &UpdateContext{Session: ss, Flow: f}
	if f.Type == flow.TypeAPI || x.IsJSONRequest(r) {
		return c, nil
	}

	if _, err := d.ContinuityManager().Continue(r.Context(), w, r, name, ContinuityOptions(payload, ss.Identity)...); err == nil {
		if payload.GetFlowID() == f.ID {
			return c, ErrContinuePreviousAction
		}
		d.Logger().
			WithField("package", pkgName).
			WithField("stack_trace", string(debug.Stack())).
			WithField("expected_request_id", payload.GetFlowID()).
			WithField("actual_request_id", f.ID).
			Debug("Flow ID from continuity manager does not match Flow ID from request.")
		return c, nil
	} else if !errors.Is(err, &continuity.ErrNotResumable) {
		return new(UpdateContext), err
	}

	return c, nil
}

func ContinuityOptions(p interface{}, i *identity.Identity) []continuity.ManagerOption {
	return []continuity.ManagerOption{
		continuity.WithPayload(p),
		continuity.WithIdentity(i),
		continuity.WithLifespan(time.Minute * 15),
	}
}

func GetFlowID(r *http.Request) (uuid.UUID, error) {
	rid := x.ParseUUID(r.URL.Query().Get("flow"))
	if rid == uuid.Nil {
		return rid, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The flow query parameter is missing or malformed."))
	}
	return rid, nil
}

func OnUnauthenticated(reg interface {
	config.Provider
	x.WriterProvider
}) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handler := session.RedirectOnUnauthenticated(reg.Config().SelfServiceFlowLoginUI(r.Context()).String())
		if x.IsJSONRequest(r) {
			handler = session.RespondWithJSONErrorOnAuthenticated(reg.Writer(), herodot.ErrUnauthorized.WithReasonf("A valid Ory Session Cookie or Ory Session Token is missing."))
		}

		handler(w, r)
	}
}
