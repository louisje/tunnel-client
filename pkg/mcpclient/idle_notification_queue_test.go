package mcpclient

import (
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIdleNotificationQueueRejectsStaleGeneration(t *testing.T) {
	q := newIdleNotificationQueue(2)
	q.reset(1)
	first := &jsonrpc.Request{Method: "notifications/tools/list_changed"}
	require.True(t, q.offerForGeneration(1, first))
	q.reset(2)
	require.Empty(t, q.drain())
	require.False(t, q.offerForGeneration(1, first))
	second := &jsonrpc.Request{Method: "notifications/resources/list_changed"}
	require.True(t, q.offerForGeneration(2, second))
	require.Equal(t, []*jsonrpc.Request{second}, q.drain())
}

func TestIdleNotificationQueueAcceptsOnlyNotifications(t *testing.T) {
	q := newIdleNotificationQueue(2)
	id, err := jsonrpc.MakeID("request")
	require.NoError(t, err)
	require.False(t, q.offer(&jsonrpc.Request{ID: id, Method: "sampling/createMessage"}))
	require.False(t, q.offer(&jsonrpc.Response{ID: id}))
	a := &jsonrpc.Request{Method: "notifications/tools/list_changed"}
	b := &jsonrpc.Request{Method: "notifications/resources/list_changed"}
	require.True(t, q.offer(a))
	require.True(t, q.offer(b))
	require.False(t, q.offer(&jsonrpc.Request{Method: "notifications/progress"}))
	require.Equal(t, []*jsonrpc.Request{a, b}, q.drain())
	require.Empty(t, q.drain())
}
