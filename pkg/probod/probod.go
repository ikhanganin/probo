// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package probod

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/getprobo/probo/pkg/awsconfig"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/crypto/passwdhash"
	"github.com/getprobo/probo/pkg/mailer"
	"github.com/getprobo/probo/pkg/probo"
	"github.com/getprobo/probo/pkg/server"
	console_v1 "github.com/getprobo/probo/pkg/server/api/console/v1"
	"github.com/getprobo/probo/pkg/usrmgr"
	"github.com/prometheus/client_golang/prometheus"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/migrator"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/unit"
	"go.opentelemetry.io/otel/trace"
)

type (
	Implm struct {
		cfg config
	}

	config struct {
		Hostname string       `json:"hostname"`
		Pg       pgConfig     `json:"pg"`
		Api      apiConfig    `json:"api"`
		Auth     authConfig   `json:"auth"`
		AWS      awsConfig    `json:"aws"`
		Mailer   mailerConfig `json:"mailer"`
	}
)

var (
	_ unit.Configurable = (*Implm)(nil)
	_ unit.Runnable     = (*Implm)(nil)
)

func New() *Implm {
	return &Implm{
		cfg: config{
			Hostname: "localhost:8080",
			Api: apiConfig{
				Addr: "localhost:8080",
			},
			Pg: pgConfig{
				Addr:     "localhost:5432",
				Username: "probod",
				Password: "probod",
				Database: "probod",
				PoolSize: 100,
			},
			Auth: authConfig{
				Password: passwordConfig{
					Pepper:     "this-is-a-secure-pepper-for-password-hashing-at-least-32-bytes",
					Iterations: 1000000,
				},
				Cookie: cookieConfig{
					Name:     "SSID",
					Secret:   "this-is-a-secure-secret-for-cookie-signing-at-least-32-bytes",
					Duration: 24,
					Domain:   "localhost",
				},
			},
			AWS: awsConfig{
				Region: "us-east-1",
				Bucket: "probod",
			},
			Mailer: mailerConfig{
				SenderEmail: "no-reply@notification.getprobo.com",
				SenderName:  "Probo",
				SMTP: smtpConfig{
					Addr: "localhost:1025",
				},
			},
		},
	}
}

func (impl *Implm) GetConfiguration() any {
	return &impl.cfg
}

func (impl *Implm) Run(
	parentCtx context.Context,
	l *log.Logger,
	r prometheus.Registerer,
	tp trace.TracerProvider,
) error {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancelCause(parentCtx)
	defer cancel(context.Canceled)

	pgClient, err := pg.NewClient(
		impl.cfg.Pg.Options(
			pg.WithLogger(l),
			pg.WithRegisterer(r),
			pg.WithTracerProvider(tp),
		)...,
	)
	if err != nil {
		return fmt.Errorf("cannot create pg client: %w", err)
	}

	pepper, err := impl.cfg.Auth.GetPepperBytes()
	if err != nil {
		return fmt.Errorf("cannot get pepper bytes: %w", err)
	}

	// Validate cookie secret
	_, err = impl.cfg.Auth.GetCookieSecretBytes()
	if err != nil {
		return fmt.Errorf("cannot get cookie secret bytes: %w", err)
	}

	awsConfig := awsconfig.NewConfig(
		l,
		httpclient.DefaultPooledClient(
			httpclient.WithLogger(l),
			httpclient.WithTracerProvider(tp),
			httpclient.WithRegisterer(r),
		),
		awsconfig.Options{
			Region:          impl.cfg.AWS.Region,
			AccessKeyID:     impl.cfg.AWS.AccessKeyID,
			SecretAccessKey: impl.cfg.AWS.SecretAccessKey,
			Endpoint:        impl.cfg.AWS.Endpoint,
		},
	)

	s3Client := s3.NewFromConfig(awsConfig)

	err = migrator.NewMigrator(pgClient, coredata.Migrations).Run(ctx, "migrations")
	if err != nil {
		return fmt.Errorf("cannot migrate database schema: %w", err)
	}

	hp, err := passwdhash.NewProfile(pepper, uint32(impl.cfg.Auth.Password.Iterations))
	if err != nil {
		return fmt.Errorf("cannot create hashing profile: %w", err)
	}

	usrmgrService, err := usrmgr.NewService(
		ctx,
		pgClient,
		hp,
		impl.cfg.Auth.Cookie.Secret,
		impl.cfg.Hostname,
	)
	if err != nil {
		return fmt.Errorf("cannot create usrmgr service: %w", err)
	}

	proboService, err := probo.NewService(ctx, pgClient, s3Client, impl.cfg.AWS.Bucket)
	if err != nil {
		return fmt.Errorf("cannot create probo service: %w", err)
	}

	serverHandler, err := server.NewServer(
		server.Config{
			AllowedOrigins: impl.cfg.Api.Cors.AllowedOrigins,
			Probo:          proboService,
			Usrmgr:         usrmgrService,
			Auth: console_v1.AuthConfig{
				CookieName:      impl.cfg.Auth.Cookie.Name,
				CookieDomain:    impl.cfg.Auth.Cookie.Domain,
				SessionDuration: time.Duration(impl.cfg.Auth.Cookie.Duration) * time.Hour,
				CookieSecret:    impl.cfg.Auth.Cookie.Secret,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("cannot create server: %w", err)
	}

	apiServerCtx, stopApiServer := context.WithCancel(context.Background())
	defer stopApiServer()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := impl.runApiServer(apiServerCtx, l, r, tp, serverHandler); err != nil {
			cancel(fmt.Errorf("api server crashed: %w", err))
		}
	}()

	mailerCtx, stopMailer := context.WithCancel(context.Background())
	mailer := mailer.NewMailer(pgClient, l, mailer.Config{
		SenderEmail: impl.cfg.Mailer.SenderEmail,
		SenderName:  impl.cfg.Mailer.SenderName,
		Addr:        impl.cfg.Mailer.SMTP.Addr,
		User:        impl.cfg.Mailer.SMTP.User,
		Password:    impl.cfg.Mailer.SMTP.Password,
		TLSRequired: impl.cfg.Mailer.SMTP.TLSRequired,
		Timeout:     time.Second * 10,
	})
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := mailer.Run(mailerCtx); err != nil {
			cancel(fmt.Errorf("mailer crashed: %w", err))
		}
	}()

	<-ctx.Done()

	stopMailer()
	stopApiServer()

	wg.Wait()

	return context.Cause(ctx)
}

func (impl *Implm) runApiServer(
	ctx context.Context,
	l *log.Logger,
	r prometheus.Registerer,
	tp trace.TracerProvider,
	handler http.Handler,
) error {
	apiServer := httpserver.NewServer(
		impl.cfg.Api.Addr,
		handler,
		httpserver.WithLogger(l),
		httpserver.WithRegisterer(r),
		httpserver.WithTracerProvider(tp),
	)

	l.Info("starting api server", log.String("addr", apiServer.Addr))

	listener, err := net.Listen("tcp", apiServer.Addr)
	if err != nil {
		return fmt.Errorf("cannot listen on %q: %w", apiServer.Addr, err)
	}
	defer listener.Close()

	serverErrCh := make(chan error, 1)
	go func() {
		err := apiServer.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- fmt.Errorf("cannot server http request: %w", err)
		}
		close(serverErrCh)
	}()

	l.Info("api server started")

	select {
	case err := <-serverErrCh:
		return err
	case <-ctx.Done():
	}

	l.InfoCtx(ctx, "shutting down api server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("cannot shutdown api server: %w", err)
	}

	return ctx.Err()
}
