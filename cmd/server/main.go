package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ysomad/gigabg/game/catalog"
	"github.com/ysomad/gigabg/lobby"
	"github.com/ysomad/gigabg/pkg/httpserver"
	"github.com/ysomad/gigabg/server"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	port := flag.Int("port", 8080, "server port")
	flag.Parse()

	ctx, notifyCancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
	defer notifyCancel()

	cards, err := catalog.New()
	if err != nil {
		return fmt.Errorf("card catalog: %w", err)
	}

	memstore := lobby.NewMemoryStore()

	testLobby, err := lobby.New(cards, 2)
	if err != nil {
		return err
	}
	testLobby.SetID("1")

	if err := memstore.CreateLobby(testLobby); err != nil {
		return err
	}

	gameServer := server.New(memstore, cards)

	srv := httpserver.New(ctx, gameServer, httpserver.WithPort(*port))

	select {
	case err := <-srv.Notify():
		slog.ErrorContext(ctx, "httpserver: "+err.Error())
	case <-ctx.Done():
		slog.InfoContext(ctx, "root context done")
	}

	if err := srv.Shutdown(ctx); err != nil {
		slog.WarnContext(ctx, "httpserver: shutdown: "+err.Error())
	}

	return nil
}
