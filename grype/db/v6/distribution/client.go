package distribution

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/spf13/afero"
	"github.com/wagoodman/go-progress"

	"github.com/anchore/clio"
	v6 "github.com/anchore/grype/grype/db/v6"
	"github.com/anchore/grype/internal/bus"
	"github.com/anchore/grype/internal/file"
	"github.com/anchore/grype/internal/log"
)

type Config struct {
	ID clio.Identification

	// check/fetch parameters
	LatestURL string
	CACert    string

	// validations
	RequireUpdateCheck bool

	// timeouts
	CheckTimeout  time.Duration
	UpdateTimeout time.Duration
}

type Client interface {
	Latest() (*LatestDocument, error)
	IsUpdateAvailable(current *v6.Description) (*Archive, error)
	ResolveArchiveURL(archive Archive) (string, error)
	Download(url, dest string, downloadProgress *progress.Manual) (string, error)
}

type client struct {
	fs                afero.Fs
	dbDownloader      file.Getter
	listingDownloader file.Getter
	config            Config
}

func DefaultConfig() Config {
	return Config{
		LatestURL:          "https://grype.anchore.io/databases",
		RequireUpdateCheck: false,
		CheckTimeout:       30 * time.Second,
		UpdateTimeout:      300 * time.Second,
	}
}

func NewClient(cfg Config) (Client, error) {
	fs := afero.NewOsFs()
	latestClient, err := defaultHTTPClient(fs, cfg.CACert, withClientTimeout(cfg.CheckTimeout), withUserAgent(cfg.ID))
	if err != nil {
		return client{}, err
	}

	dbClient, err := defaultHTTPClient(fs, cfg.CACert, withClientTimeout(cfg.UpdateTimeout), withUserAgent(cfg.ID))
	if err != nil {
		return client{}, err
	}

	return client{
		fs:                fs,
		listingDownloader: file.NewGetter(cfg.ID, latestClient),
		dbDownloader:      file.NewGetter(cfg.ID, dbClient),
		config:            cfg,
	}, nil
}

// IsUpdateAvailable indicates if there is a new update available as a boolean, and returns the latest db information
// available for this schema.
func (c client) IsUpdateAvailable(current *v6.Description) (*Archive, error) {
	log.Debugf("checking for available database updates")

	latestDoc, err := c.Latest()
	if err != nil {
		if c.config.RequireUpdateCheck {
			return nil, fmt.Errorf("check for vulnerability database update failed: %+v", err)
		}
		log.Warnf("unable to check for vulnerability database update")
		log.Debugf("check for vulnerability update failed: %+v", err)
	}

	archive, message := c.isUpdateAvailable(current, latestDoc)

	if message != "" {
		log.Warn(message)
		bus.Notify(message)
	}

	return archive, err
}

func (c client) isUpdateAvailable(current *v6.Description, candidate *LatestDocument) (*Archive, string) {
	if candidate == nil {
		return nil, ""
	}

	var message string
	switch candidate.Status {
	case StatusDeprecated:
		message = "this version of grype will soon stop receiving vulnerability database updates, please update grype"
	case StatusEndOfLife:
		message = "this version of grype is no longer receiving vulnerability database updates, please update grype"
	}

	// compare created data to current db date
	if isSupersededBy(current, candidate.Description) {
		log.Debugf("database update available: %s", candidate.Description)
		return &candidate.Archive, message
	}

	log.Debugf("no database update available")
	return nil, message
}

func (c client) ResolveArchiveURL(archive Archive) (string, error) {
	// download the db to the temp dir
	u, err := url.Parse(c.latestURL())
	if err != nil {
		return "", fmt.Errorf("unable to parse db URL %q: %w", c.latestURL(), err)
	}

	u.Path = path.Join(path.Dir(u.Path), path.Clean(archive.Path))

	// from go-getter, adding a checksum as a query string will validate the payload after download
	// note: the checksum query parameter is not sent to the server
	query := u.Query()
	if archive.Checksum != "" {
		query.Add("checksum", archive.Checksum)
	}
	u.RawQuery = query.Encode()

	return u.String(), nil
}

func (c client) Download(archiveURL, dest string, downloadProgress *progress.Manual) (string, error) {
	defer downloadProgress.SetCompleted()

	if err := os.MkdirAll(dest, 0700); err != nil {
		return "", fmt.Errorf("unable to create db download root dir: %w", err)
	}

	// note: as much as I'd like to use the afero FS abstraction here, the go-getter library does not support it
	tempDir, err := os.MkdirTemp(dest, "grype-db-download")
	if err != nil {
		return "", fmt.Errorf("unable to create db client temp dir: %w", err)
	}

	// go-getter will automatically extract all files within the archive to the temp dir
	err = c.dbDownloader.GetToDir(tempDir, archiveURL, downloadProgress)
	if err != nil {
		removeAllOrLog(afero.NewOsFs(), tempDir)
		return "", fmt.Errorf("unable to download db: %w", err)
	}

	return tempDir, nil
}

// Latest loads a LatestDocument from the configured URL.
func (c client) Latest() (*LatestDocument, error) {
	tempFile, err := afero.TempFile(c.fs, "", "grype-db-listing")
	if err != nil {
		return nil, fmt.Errorf("unable to create listing temp file: %w", err)
	}
	defer func() {
		log.CloseAndLogError(tempFile, tempFile.Name())
		err := c.fs.RemoveAll(tempFile.Name())
		if err != nil {
			log.WithFields("error", err, "file", tempFile.Name()).Errorf("failed to remove file")
		}
	}()

	err = c.listingDownloader.GetFile(tempFile.Name(), c.latestURL())
	if err != nil {
		return nil, fmt.Errorf("unable to download listing: %w", err)
	}

	return NewLatestFromFile(c.fs, tempFile.Name())
}

func (c client) latestURL() string {
	u := c.config.LatestURL
	// allow path to be specified directly to a json file, or the path without version information
	if !strings.HasSuffix(u, ".json") {
		u = strings.TrimRight(u, "/")
		u = fmt.Sprintf("%s/v%d/%s", u, v6.ModelVersion, LatestFileName)
	}
	return u
}

func withClientTimeout(timeout time.Duration) func(*http.Client) {
	return func(c *http.Client) {
		c.Timeout = timeout
	}
}

func withUserAgent(id clio.Identification) func(*http.Client) {
	return func(c *http.Client) {
		*(c) = *newHTTPClientWithDefaultUserAgent(c.Transport, fmt.Sprintf("%s %s", id.Name, id.Version))
	}
}

func defaultHTTPClient(fs afero.Fs, caCertPath string, postProcessor ...func(*http.Client)) (*http.Client, error) {
	httpClient := cleanhttp.DefaultClient()
	httpClient.Timeout = 30 * time.Second
	if caCertPath != "" {
		rootCAs := x509.NewCertPool()

		pemBytes, err := afero.ReadFile(fs, caCertPath)
		if err != nil {
			return nil, fmt.Errorf("unable to configure root CAs for curator: %w", err)
		}
		rootCAs.AppendCertsFromPEM(pemBytes)

		httpClient.Transport.(*http.Transport).TLSClientConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			RootCAs:    rootCAs,
		}
	}

	for _, pp := range postProcessor {
		pp(httpClient)
	}

	return httpClient, nil
}

func removeAllOrLog(fs afero.Fs, dir string) {
	if err := fs.RemoveAll(dir); err != nil {
		log.WithFields("error", err).Warnf("failed to remove path %q", dir)
	}
}

func isSupersededBy(current *v6.Description, candidate v6.Description) bool {
	if current == nil {
		log.Debug("cannot find existing metadata, using update...")
		// any valid update beats no database, use it!
		return true
	}

	if !current.SchemaVersion.Valid() {
		log.Error("existing database has no schema version, doing nothing...")
		return false
	}

	if !candidate.SchemaVersion.Valid() {
		log.Error("update has no schema version, doing nothing...")
		return false
	}

	if candidate.SchemaVersion.Model != current.SchemaVersion.Model {
		log.WithFields("want", current.SchemaVersion.Model, "received", candidate.SchemaVersion.Model).Warn("update is for a different DB schema, skipping...")
		return false
	}

	if candidate.Built.After(current.Built.Time) {
		d := candidate.Built.Sub(current.Built.Time).String()
		log.WithFields("existing", current.Built.String(), "candidate", candidate.Built.String(), "delta", d).Debug("existing database is older than candidate update, using update...")
		// the listing is newer than the existing db, use it!
		return true
	}

	log.Debugf("existing database is already up to date")
	return false
}

func newHTTPClientWithDefaultUserAgent(baseTransport http.RoundTripper, userAgent string) *http.Client {
	return &http.Client{
		Transport: roundTripperWithUserAgent{
			transport: baseTransport,
			userAgent: userAgent,
		},
	}
}

type roundTripperWithUserAgent struct {
	transport http.RoundTripper
	userAgent string
}

func (r roundTripperWithUserAgent) RoundTrip(req *http.Request) (*http.Response, error) {
	clonedReq := req.Clone(req.Context())

	if clonedReq.Header.Get("User-Agent") == "" {
		clonedReq.Header.Set("User-Agent", r.userAgent)
	}

	return r.transport.RoundTrip(clonedReq)
}
