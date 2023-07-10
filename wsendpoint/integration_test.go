package wsendpoint_test

import (
	"context"
	"net"
	"net/http"
	"path"
	"testing"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/redsift/go-stats/wsendpoint"
	"github.com/stretchr/testify/require"
)

type msg struct {
	data []byte
	err  error
}

func TestWebsocket(t *testing.T) {
	h, c := wsendpoint.New(context.TODO())
	server := &http.Server{
		Handler: h,
	}

	tmpdir := t.TempDir()
	sock := path.Join(tmpdir, "http.sock")
	l, err := net.Listen("unix", sock)
	require.NoError(t, err, "http listener failed")

	go func() {
		_ = server.Serve(l)
	}()
	defer server.Close()

	d := ws.Dialer{
		NetDial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial("unix", sock)
		},
	}

	conn, _, _, err := d.Dial(context.TODO(), "ws://localhost:1234")
	require.NoError(t, err, "ws client failed")
	defer conn.Close()

	res := make(chan msg, 1)
	go func() {
		defer close(res)
		for round := 0; ; round++ {
			data, err := wsutil.ReadServerText(conn)
			res <- msg{data, err}
		}
	}()

	require.Empty(t, res)

	// ---
	c.Inform("test", "bla")
	msg := <-res

	require.NoError(t, msg.err)
	require.Equal(t, `{"event":{"level":"info","title":"test","text":"bla"}}`, string(msg.data))

	// ---
	c.Inform("test", "bla", "some:tag")
	msg = <-res

	require.NoError(t, msg.err)
	require.Equal(t, `{"event":{"level":"info","title":"test","text":"bla","tags":{"some":"tag"}}}`, string(msg.data))

	// ---
	c.With("with:value").Inform("test", "bla", "some:tag")
	msg = <-res

	require.NoError(t, msg.err)
	require.Equal(t, `{"event":{"level":"info","title":"test","text":"bla","tags":{"with":"value","some":"tag"}}}`, string(msg.data))
}
