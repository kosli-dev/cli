// Package kosli is the only way a command talks to Kosli.
//
// --host and --api-token may each be a comma-separated list of the same
// length, naming several Kosli servers. A command reads from the first, the
// primary, and sends what it gathered once to every one of them at the same
// time, so every server records the same evidence. API tokens therefore must
// not contain commas.
package kosli

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
)

// Host is one Kosli server and the API token for it.
type Host struct {
	URL    string
	Token  string
	DryRun bool
	client *requests.Client
	logger *log.Logger
	// label names the host in its log lines when there are several hosts.
	label   string
	primary bool
}

// Do sends a request to this host. The caller builds the URL from h.URL and
// sets h.Token and h.DryRun.
func (h Host) Do(p *requests.RequestParams) (*requests.HTTPResponse, error) {
	return h.client.Do(p)
}

// IsPrimary reports whether h is the host whose output is shown.
func (h Host) IsPrimary() bool {
	return h.primary
}

// Client is the HTTP client requests to this host go through, for the
// clients in other packages that build their own requests.
func (h Host) Client() *requests.Client {
	return h.client
}

// Info shows the primary's line; a secondary's only shows with --debug.
func (h Host) Info(format string, v ...any) {
	if h.primary {
		h.logger.Info(format, v...)
		return
	}
	h.logger.Debug(h.label+format, v...)
}

// Warn shows for every host, so a secondary's lost evidence is not silent.
func (h Host) Warn(format string, v ...any) {
	h.logger.Warn(h.label+format, v...)
}

// Kosli is every host a command talks to, the primary first.
type Kosli struct {
	hosts []Host
}

// New reads hosts and tokens as comma-separated lists. Lists of different
// lengths, or a single host, are taken whole as one host. A DRY_RUN token puts
// its own host in dry-run.
func New(client *requests.Client, logger *log.Logger, hosts, tokens string, dryRun bool) *Kosli {
	hostList, tokenList := strings.Split(hosts, ","), strings.Split(tokens, ",")
	if !HasMultipleHosts(hostList, tokenList) {
		hostList, tokenList = []string{hosts}, []string{tokens}
	}
	k := &Kosli{}
	for i := range hostList {
		h := Host{
			URL:     hostList[i],
			Token:   tokenList[i],
			DryRun:  dryRun || tokenList[i] == "DRY_RUN",
			client:  client,
			logger:  logger,
			primary: i == 0,
		}
		if len(hostList) > 1 {
			h.label = "[" + h.URL + "] "
		}
		k.hosts = append(k.hosts, h)
	}
	return k
}

func HasMultipleHosts(hosts, tokens []string) bool {
	return len(hosts) > 1 && len(tokens) > 1 && len(hosts) == len(tokens)
}

// Query is the host to read from: the primary. Reading every host could give
// each a different answer, and only one can be shown.
func (k *Kosli) Query() Host {
	return k.hosts[0]
}

// Hosts lists every host, the primary first.
func (k *Kosli) Hosts() []Host {
	return k.hosts
}

// Send runs pipeline on every host at the same time. A failure on one host
// does not stop the others.
func (k *Kosli) Send(pipeline func(h Host) error) error {
	_, err := Each(k, func(h Host) (struct{}, error) { return struct{}{}, pipeline(h) })
	return err
}

// Each runs pipeline on every host at the same time and returns each host's
// result at its index, the primary's first.
func Each[T any](k *Kosli, pipeline func(h Host) (T, error)) ([]T, error) {
	results := make([]T, len(k.hosts))
	errs := make([]error, len(k.hosts))
	var wg sync.WaitGroup
	for i, h := range k.hosts {
		wg.Go(func() { results[i], errs[i] = pipeline(h) })
	}
	wg.Wait()
	return results, k.join(errs)
}

// join names the host of each failure once there is more than one host.
func (k *Kosli) join(errs []error) error {
	if len(k.hosts) == 1 {
		return errs[0]
	}
	var failures []string
	for i, err := range errs {
		if err != nil {
			failures = append(failures, fmt.Sprintf("[%s]\n%s", k.hosts[i].URL, err))
		}
	}
	if len(failures) == 0 {
		return nil
	}
	return errors.New(strings.Join(failures, "\n"))
}
