// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/persistence/sql/update"
	"github.com/ory/kratos/x"
	"github.com/ory/pop/v6"
	"github.com/ory/x/otelx"
	"github.com/ory/x/pagination/keysetpagination"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/uuidx"
)

var _ courier.Persister = new(Persister)

func (p *Persister) AddMessage(ctx context.Context, m *courier.Message) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.AddMessage")
	defer otelx.End(span, &err)

	m.NID = p.NetworkID(ctx)
	m.Status = courier.MessageStatusQueued
	return sqlcon.HandleError(p.GetConnection(ctx).Create(m)) // do not create eager to avoid identity injection.
}

func (p *Persister) ListMessages(ctx context.Context, filter courier.ListCourierMessagesParameters, opts []keysetpagination.Option) (_ []courier.Message, _ int64, _ *keysetpagination.Paginator, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ListMessages")
	defer otelx.End(span, &err)

	q := p.GetConnection(ctx).Where("nid=?", p.NetworkID(ctx))

	if filter.Status != nil {
		q = q.Where("status=?", *filter.Status)
	}

	if filter.Recipient != "" {
		q = q.Where("recipient=?", filter.Recipient)
	}

	count, err := q.Count(&courier.Message{})
	if err != nil {
		return nil, 0, nil, sqlcon.HandleError(err)
	}

	opts = append(opts, keysetpagination.WithDefaultToken(new(courier.Message).DefaultPageToken()))
	opts = append(opts, keysetpagination.WithDefaultSize(10))
	opts = append(opts, keysetpagination.WithColumn("created_at", "DESC"))
	paginator := keysetpagination.GetPaginator(opts...)

	if _, err := uuid.FromString(paginator.Token().Parse("id")["id"]); err != nil {
		return nil, 0, nil, errors.WithStack(x.PageTokenInvalid)
	}

	messages := make([]courier.Message, paginator.Size())
	if err := q.Scope(keysetpagination.Paginate[courier.Message](paginator)).
		All(&messages); err != nil {
		return nil, 0, nil, sqlcon.HandleError(err)
	}

	messages, nextPage := keysetpagination.Result(messages, paginator)
	return messages, int64(count), nextPage, nil
}

func (p *Persister) NextMessages(ctx context.Context, limit uint8) (messages []courier.Message, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.NextMessages")
	defer otelx.End(span, &err)

	if err := p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		var m []courier.Message
		if err := tx.
			Where("nid = ? AND status = ?",
				p.NetworkID(ctx),
				courier.MessageStatusQueued,
			).
			Order("created_at ASC").
			Limit(int(limit)).
			All(&m); err != nil {
			return err
		}

		if len(m) == 0 {
			return sql.ErrNoRows
		}

		for i := range m {
			message := &m[i]
			message.Status = courier.MessageStatusProcessing
			if err := update.Generic(ctx, p.GetConnection(ctx), p.r.Tracer(ctx).Tracer(), message, "status"); err != nil {
				return err
			}
		}

		messages = m
		return nil
	}); err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, errors.WithStack(courier.ErrQueueEmpty)
		}
		return nil, sqlcon.HandleError(err)
	}

	if len(messages) == 0 {
		return nil, errors.WithStack(courier.ErrQueueEmpty)
	}

	return messages, nil
}

func (p *Persister) LatestQueuedMessage(ctx context.Context) (_ *courier.Message, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.LatestQueuedMessage")
	defer otelx.End(span, &err)

	var m courier.Message
	if err := p.GetConnection(ctx).
		Where("nid = ? AND status = ?",
			p.NetworkID(ctx),
			courier.MessageStatusQueued,
		).
		Order("created_at DESC").
		First(&m); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(courier.ErrQueueEmpty)
		}
		return nil, sqlcon.HandleError(err)
	}

	return &m, nil
}

func (p *Persister) SetMessageStatus(ctx context.Context, id uuid.UUID, ms courier.MessageStatus) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.SetMessageStatus")
	defer otelx.End(span, &err)

	count, err := p.GetConnection(ctx).RawQuery(
		"UPDATE courier_messages SET status = ? WHERE id = ? AND nid = ?",
		ms,
		id,
		p.NetworkID(ctx),
	).ExecWithCount()
	if err != nil {
		return sqlcon.HandleError(err)
	}

	if count == 0 {
		return errors.WithStack(sqlcon.ErrNoRows)
	}

	return nil
}

func (p *Persister) IncrementMessageSendCount(ctx context.Context, id uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.IncrementMessageSendCount")
	defer otelx.End(span, &err)

	count, err := p.GetConnection(ctx).RawQuery(
		"UPDATE courier_messages SET send_count = send_count + 1 WHERE id = ? AND nid = ?",
		id,
		p.NetworkID(ctx),
	).ExecWithCount()
	if err != nil {
		return sqlcon.HandleError(err)
	}

	if count == 0 {
		return errors.WithStack(sqlcon.ErrNoRows)
	}

	return nil
}

func (p *Persister) FetchMessage(ctx context.Context, msgID uuid.UUID) (_ *courier.Message, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.FetchMessage")
	defer otelx.End(span, &err)

	var message courier.Message
	if err := p.GetConnection(ctx).
		Where("id = ? AND nid = ?", msgID, p.NetworkID(ctx)).
		Eager().
		First(&message); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &message, nil
}

func (p *Persister) RecordDispatch(ctx context.Context, msgID uuid.UUID, status courier.CourierMessageDispatchStatus, err error) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.RecordDispatch")
	defer otelx.End(span, &err)

	dispatch := courier.MessageDispatch{
		ID:        uuidx.NewV4(),
		MessageID: msgID,
		Status:    status,
		NID:       p.NetworkID(ctx),
	}

	if err != nil {
		// We use herodot as a carrier for the error's data
		her := herodot.ToDefaultError(err, "")
		content, mErr := json.Marshal(her)
		if mErr != nil {
			return errors.WithStack(mErr)
		}
		dispatch.Error = content
	}

	if err := p.GetConnection(ctx).Create(&dispatch); err != nil {
		return sqlcon.HandleError(err)
	}

	return nil
}
