package gdpr_test

import (
	"context"
	"testing"
	"unsafe"

	"azugo.io/azugo"
	"github.com/go-quicktest/qt"

	"github.com/gmb-lib/go-gdpr-audit/gdpr"
	"github.com/gmb-lib/go-platform-kit/broker"
	"github.com/gmb-lib/go-platform-kit/propagation"
)

// overBuffer returns a string that shares buf's bytes, as a value read from a
// request shares the buffer the server reuses for the next request.
func overBuffer(buf []byte) string {
	return unsafe.String(&buf[0], len(buf)) //nolint:gosec // the aliasing a request read makes, on purpose
}

// requestBuffers holds one buffer per request-read value a record carries.
type requestBuffers struct {
	correlation, actor, subject, resource, recipient []byte
}

func newRequestBuffers() *requestBuffers {
	return &requestBuffers{
		correlation: []byte("corr-1"),
		actor:       []byte("user-1"),
		subject:     []byte("subject-1"),
		resource:    []byte("doc-1"),
		recipient:   []byte("first"),
	}
}

// record records an access whose values all point into b, as a handler's do.
func (b *requestBuffers) record(t *testing.T, client *gdpr.Client) error {
	t.Helper()

	var rerr error

	withCtx(t, func(ctx *azugo.Context) {
		ctx.SetUserValue(propagation.RequestValueName(), overBuffer(b.correlation))

		rerr = client.Record(ctx, gdpr.EventDocumentAccess, gdpr.Access{
			Actor:        broker.Actor{ID: overBuffer(b.actor), Type: "user"},
			DataSubjects: []string{overBuffer(b.subject)},
			Resource:     broker.Resource{Type: gdpr.ResourceDocument, ID: overBuffer(b.resource)},
			Operation:    broker.OpRead,
			LawfulBasis:  gdpr.BasisContract,
			Attributes:   map[string]any{"recipient": overBuffer(b.recipient)},
		})
	})

	return rerr
}

// reuse writes the next request into every buffer.
func (b *requestBuffers) reuse() {
	copy(b.correlation, "corr-2")
	copy(b.actor, "user-2")
	copy(b.subject, "subject-2")
	copy(b.resource, "doc-2")
	copy(b.recipient, "later")
}

func checkFirstRequest(t *testing.T, rec *broker.Envelope) {
	t.Helper()

	qt.Check(t, qt.Equals(rec.CorrelationID, "corr-1"))
	qt.Assert(t, qt.IsNotNil(rec.Actor))
	qt.Check(t, qt.Equals(rec.Actor.ID, "user-1"))
	qt.Assert(t, qt.Equals(len(rec.DataSubjects), 1))
	qt.Check(t, qt.Equals(rec.DataSubjects[0], "subject-1"))
	qt.Assert(t, qt.IsNotNil(rec.Resource))
	qt.Check(t, qt.Equals(rec.Resource.ID, "doc-1"))
	qt.Check(t, qt.Equals(str(rec.Attributes["recipient"]), "first"))
}

// A record the access audit could not take is queued and posted later, after
// the server has reused the request's buffer, so the queue must hold copies of
// its own: else the posted record names whatever the next request carried.
func TestBufferedRecordKeepsItsOwnCopyOfTheRequestsValues(t *testing.T) {
	p := &fakePoster{failUntil: 1}
	ob := gdpr.NewMemoryOutbox(16)
	client, err := gdpr.New(testConfig(), p, gdpr.Options{Outbox: ob})
	qt.Assert(t, qt.IsNil(err))

	bufs := newRequestBuffers()
	qt.Assert(t, qt.IsNil(bufs.record(t, client)))
	qt.Assert(t, qt.Equals(ob.Len(), 1))

	bufs.reuse()

	qt.Assert(t, qt.IsNil(client.Flush(context.Background())))

	recs := p.accepted()
	qt.Assert(t, qt.Equals(len(recs), 1))
	checkFirstRequest(t, recs[0])
}

// fullOutbox refuses every record, so the client dead-letters it at once.
type fullOutbox struct{}

func (fullOutbox) Enqueue(*broker.Envelope) error { return gdpr.ErrOutboxFull }

func (fullOutbox) Dequeue(ctx context.Context) (*broker.Envelope, error) {
	<-ctx.Done()

	return nil, ctx.Err()
}

func (fullOutbox) Len() int { return 0 }

// A dead-letter sink may keep what it is handed past the request, so it gets
// the same copies the queue does.
func TestDeadLetteredRecordKeepsItsOwnCopyOfTheRequestsValues(t *testing.T) {
	var dead []*broker.Envelope

	p := &fakePoster{alwaysFail: true}
	client, err := gdpr.New(testConfig(), p, gdpr.Options{
		Outbox:     fullOutbox{},
		DeadLetter: func(rec *broker.Envelope) { dead = append(dead, rec) },
	})
	qt.Assert(t, qt.IsNil(err))

	bufs := newRequestBuffers()
	qt.Assert(t, qt.ErrorIs(bufs.record(t, client), gdpr.ErrOutboxFull))
	qt.Assert(t, qt.Equals(len(dead), 1))

	bufs.reuse()

	checkFirstRequest(t, dead[0])
}
