package kosli

import (
	"bytes"
	"errors"
	"sync"
	"testing"
	"time"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestKosli(t *testing.T, hosts, tokens string, debug bool) (*Kosli, *bytes.Buffer) {
	t.Helper()
	var out bytes.Buffer
	l := log.NewLogger(&out, &out, debug)
	client, err := requests.NewKosliClient("", 0, false, l)
	require.NoError(t, err)
	return New(client, l, hosts, tokens, false), &out
}

func urls(k *Kosli) []string {
	var got []string
	for _, h := range k.Hosts() {
		got = append(got, h.URL)
	}
	return got
}

func TestNewSplitsMatchingListsIntoHosts(t *testing.T) {
	for _, tt := range []struct {
		name, hosts, tokens string
		want                []string
	}{
		{"two hosts and two tokens", "http://a,http://b", "one,two", []string{"http://a", "http://b"}},
		{"three hosts and three tokens", "http://a,http://b,http://c", "1,2,3", []string{"http://a", "http://b", "http://c"}},
		{"one host and two tokens", "http://a", "one,two", []string{"http://a"}},
		{"three hosts and two tokens", "http://a,http://b,http://c", "one,two", []string{"http://a,http://b,http://c"}},
		{"two hosts and one token", "http://a,http://b", "one", []string{"http://a,http://b"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			k, _ := newTestKosli(t, tt.hosts, tt.tokens, false)
			assert.Equal(t, tt.want, urls(k))
		})
	}
}

func TestNewPutsADryRunTokensHostInDryRun(t *testing.T) {
	k, _ := newTestKosli(t, "http://a,http://b", "one,DRY_RUN", false)

	assert.False(t, k.Hosts()[0].DryRun)
	assert.True(t, k.Hosts()[1].DryRun)
}

func TestQueryIsThePrimary(t *testing.T) {
	k, _ := newTestKosli(t, "http://a,http://b", "one,two", false)

	assert.Equal(t, "http://a", k.Query().URL)
	assert.True(t, k.Query().IsPrimary())
	assert.False(t, k.Hosts()[1].IsPrimary())
}

func TestSendRunsEveryHostsPipelineAtOnce(t *testing.T) {
	k, _ := newTestKosli(t, "http://a,http://b", "one,two", false)
	var started sync.WaitGroup
	started.Add(2)
	allStarted := make(chan struct{})
	go func() { started.Wait(); close(allStarted) }()

	err := k.Send(func(Host) error {
		started.Done()
		select {
		case <-allStarted:
			return nil
		case <-time.After(5 * time.Second):
			return errors.New("the other host's pipeline did not start until this one finished")
		}
	})

	assert.NoError(t, err)
}

func TestEachReturnsEveryHostsResultInHostOrder(t *testing.T) {
	k, _ := newTestKosli(t, "http://a,http://b", "one,two", false)

	results, err := Each(k, func(h Host) (string, error) { return h.Token, nil })

	require.NoError(t, err)
	assert.Equal(t, []string{"one", "two"}, results)
}

func TestSendNamesEachFailingHost(t *testing.T) {
	k, _ := newTestKosli(t, "http://a,http://b", "one,two", false)

	err := k.Send(func(h Host) error {
		if h.IsPrimary() {
			return nil
		}
		return errors.New("boom")
	})

	assert.EqualError(t, err, "[http://b]\nboom")
}

func TestSendLeavesASingleHostsErrorAlone(t *testing.T) {
	k, _ := newTestKosli(t, "http://a", "one", false)

	err := k.Send(func(Host) error { return errors.New("boom") })

	assert.EqualError(t, err, "boom")
}

func TestSecondaryInfoOnlyShowsInDebug(t *testing.T) {
	k, out := newTestKosli(t, "http://a,http://b", "one,two", false)

	_ = k.Send(func(h Host) error { h.Info("sent"); return nil })

	assert.Equal(t, "sent\n", out.String())
}

func TestSecondaryInfoIsLabelledInDebug(t *testing.T) {
	k, out := newTestKosli(t, "http://a,http://b", "one,two", true)

	k.Hosts()[1].Info("sent")

	assert.Equal(t, "[debug] [http://b] sent\n", out.String())
}

func TestSecondaryWarningShowsLabelled(t *testing.T) {
	k, out := newTestKosli(t, "http://a,http://b", "one,two", false)

	k.Hosts()[1].Warn("not recorded")

	assert.Equal(t, "[warning] [http://b] not recorded\n", out.String())
}
