package mcpclient

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestExperimentServerRequestDuringToolCall(t *testing.T) {
	base := newStubSerializedForwardingConnection()
	tr := NewSerializedForwardingTransport(&stubSerializedForwardingTransport{conn: base}).(*serializedForwardingTransport)
	conn, err := tr.Connect(context.Background())
	require.NoError(t, err)
	toolID, _ := jsonrpc.MakeID("tool")
	serverID, _ := jsonrpc.MakeID("server")
	_, err = conn.Write(context.Background(), nil, &jsonrpc.Request{ID: toolID, Method: "tools/call"})
	require.NoError(t, err)
	serverReq := &jsonrpc.Request{ID: serverID, Method: "sampling/createMessage"}
	base.enqueueRead(serverReq, nil)
	msg, err := conn.Read(context.Background())
	require.NoError(t, err)
	require.Same(t, serverReq, msg)
	requireLifecycleLockHeld(t, tr)
	base.enqueueRead(&jsonrpc.Response{ID: toolID}, nil)
	_, err = conn.Read(context.Background())
	require.NoError(t, err)
	requireLifecycleLockReleased(t, tr)
}

func TestExperimentRetiredRequestDropsNotification(t *testing.T) {
	base := newStubSerializedForwardingConnection()
	tr := NewSerializedForwardingTransportWithDeadlineRetirement(&stubSerializedForwardingTransport{conn: base}).(*serializedForwardingTransport)
	a, err := tr.Connect(context.Background())
	require.NoError(t, err)
	oldID, _ := jsonrpc.MakeID("old")
	_, err = a.Write(context.Background(), nil, &jsonrpc.Request{ID: oldID, Method: "tools/call"})
	require.NoError(t, err)
	require.True(t, a.(ResponseDeadlineRetiringConnection).RetireResponseDeadline())
	b, err := tr.Connect(context.Background())
	require.NoError(t, err)
	newID, _ := jsonrpc.MakeID("new")
	_, err = b.Write(context.Background(), nil, &jsonrpc.Request{ID: newID, Method: "tools/call"})
	require.NoError(t, err)
	base.enqueueRead(&jsonrpc.Request{Method: "notifications/tools/list_changed"}, nil)
	expected := &jsonrpc.Response{ID: newID}
	base.enqueueRead(expected, nil)
	msg, err := b.Read(context.Background())
	require.NoError(t, err)
	require.Same(t, expected, msg)
}

func TestExperimentIdleNotificationCanBeReadIfReaderStaysAlive(t *testing.T) {
	base := newStubSerializedForwardingConnection()
	tr := NewSerializedForwardingTransport(&stubSerializedForwardingTransport{conn: base}).(*serializedForwardingTransport)
	conn, err := tr.Connect(context.Background())
	require.NoError(t, err)
	toolID, _ := jsonrpc.MakeID("tool")
	_, err = conn.Write(context.Background(), nil, &jsonrpc.Request{ID: toolID, Method: "tools/call"})
	require.NoError(t, err)
	base.enqueueRead(&jsonrpc.Response{ID: toolID}, nil)
	_, err = conn.Read(context.Background())
	require.NoError(t, err)
	requireLifecycleLockReleased(t, tr)
	notification := &jsonrpc.Request{Method: "notifications/tools/list_changed"}
	base.enqueueRead(notification, nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	msg, err := conn.Read(ctx)
	require.NoError(t, err)
	require.Same(t, notification, msg)
}
func TestExperimentLateServerResponseAfterToolCompletion(t *testing.T) {
	base := newStubSerializedForwardingConnection()
	tr := NewSerializedForwardingTransport(&stubSerializedForwardingTransport{conn: base}).(*serializedForwardingTransport)
	conn, err := tr.Connect(context.Background())
	require.NoError(t, err)
	toolID, _ := jsonrpc.MakeID("tool")
	serverID, _ := jsonrpc.MakeID("server")
	_, err = conn.Write(context.Background(), nil, &jsonrpc.Request{ID: toolID, Method: "tools/call"})
	require.NoError(t, err)
	base.enqueueRead(&jsonrpc.Request{ID: serverID, Method: "sampling/createMessage"}, nil)
	_, err = conn.Read(context.Background())
	require.NoError(t, err)
	base.enqueueRead(&jsonrpc.Response{ID: toolID}, nil)
	_, err = conn.Read(context.Background())
	require.NoError(t, err)
	requireLifecycleLockReleased(t, tr)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = conn.Write(ctx, nil, &jsonrpc.Response{ID: serverID})
	require.NoError(t, err)
	require.Equal(t, int32(2), base.writeCalls.Load())
}

func TestExperimentServerResponseOnSameLogicalConnection(t *testing.T) {
	base := newStubSerializedForwardingConnection()
	tr := NewSerializedForwardingTransport(&stubSerializedForwardingTransport{conn: base}).(*serializedForwardingTransport)
	conn, err := tr.Connect(context.Background())
	require.NoError(t, err)
	toolID, _ := jsonrpc.MakeID("tool")
	serverID, _ := jsonrpc.MakeID("server")
	_, err = conn.Write(context.Background(), nil, &jsonrpc.Request{ID: toolID, Method: "tools/call"})
	require.NoError(t, err)
	base.enqueueRead(&jsonrpc.Request{ID: serverID, Method: "sampling/createMessage"}, nil)
	_, err = conn.Read(context.Background())
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	_, err = conn.Write(ctx, nil, &jsonrpc.Response{ID: serverID})
	require.ErrorIs(t, err, context.DeadlineExceeded, "same logical connection response should be blocked by lifecycle lock")
}
func TestExperimentServerResponseOnSeparateLogicalConnection(t *testing.T) {
	base := newStubSerializedForwardingConnection()
	tr := NewSerializedForwardingTransport(&stubSerializedForwardingTransport{conn: base}).(*serializedForwardingTransport)
	a, err := tr.Connect(context.Background())
	require.NoError(t, err)
	toolID, _ := jsonrpc.MakeID("tool")
	serverID, _ := jsonrpc.MakeID("server")
	_, err = a.Write(context.Background(), nil, &jsonrpc.Request{ID: toolID, Method: "tools/call"})
	require.NoError(t, err)
	base.enqueueRead(&jsonrpc.Request{ID: serverID, Method: "sampling/createMessage"}, nil)
	_, err = a.Read(context.Background())
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	_, err = tr.Connect(ctx)
	require.ErrorIs(t, err, context.DeadlineExceeded, "separate logical connection Connect should be blocked by lifecycle lock")
}
